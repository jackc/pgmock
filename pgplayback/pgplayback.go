// Package pgplayback can be used to playback database communication with Postgres.
// It's built on top of pgmockproxy, which is primarily used to analyze Postgres wire protocol communication.
//
// You can install the pgmockproxy to play with it manually:
// $ go install github.com/jackc/pgmock/pgmockproxy@latest
//
// See https://github.com/jackc/pgmock for more details.
package pgplayback

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/jackc/pgmock"
	"github.com/jackc/pgmock/pgmockproxy/proxy"
	"github.com/jackc/pgproto3/v2"
	"github.com/jackc/pgx/v4"
)

// Flags should be called by package initialization to configure pgplayback
// using flags with "go test".
//
// Available flags:
// --pgplayback-bypass connects to Postgres without touching playback files
// --pgplayback-update updates playback files by listening to communication with a proxy
// --pgplayback-remote <address> sets a Postgres server address
// --pgplayback-strict exposes and stores explicit low-level communications details such as credentials
// --pgplayback-debug prints extra debugging information
//
// Example: "go test --pgplayback-update"
func Flags() *Options {
	var o Options
	flag.BoolVar(&o.Bypass, "pgplayback-bypass", false, "Bypass connects to Postgres without touching playback files")
	flag.BoolVar(&o.Update, "pgplayback-update", false, "Update playback files by listening to communication with a proxy")
	flag.StringVar(&o.Remote, "pgplayback-remote", "", "Postgres server address")
	flag.BoolVar(&o.Strict, "pgplayback-strict", false, "Stores exact wire communication without stripping credentials and other low-level details")
	flag.BoolVar(&o.Debug, "pgplayback-debug", false, "Print extra debugging information")
	return &o
}

// Options for using the transport.
type Options struct {
	// Bypass connects to Postgres without touching playback files.
	Bypass bool

	// Update playback files using the proxy to the regular database connection.
	Update bool

	// Remote server to connect when using Update.
	Remote string

	// Strict stores exact wire communication without stripping credentials and other low-level details.
	// Useful for debugging, but not recommended for regular testing.
	Strict bool

	// Debug prints extra information when executing.
	Debug bool
}

// New Transport.
func New(name string, options *Options) *Transport {
	return &Transport{
		Name:    name,
		Options: options,
	}
}

// Transport can be used to eavesdrop database traffic with pgxmock,
// and save wire protocol communication to use on tests.
//
// To try pgmockproxy manually, use:
// $ go install github.com/jackc/pgmock/pgmockproxy@latest
type Transport struct {
	// Options for using the transport.
	*Options

	// Name used to save the replay file.
	// Example: testdata/select.pgplayback
	Name string

	// conn is a connection to a Postgres database or to the replay database.
	conn *pgx.Conn

	// fields used for playback and update:
	script       *pgmock.Script
	listener     net.Listener
	proxyErrChan chan error
}

// Connect to database or replay file with a connection string.
// See pgconn.Connect for details.
func (t *Transport) Connect(ctx context.Context, connString string) (conn Postgres, err error) {
	connConfig, err := pgx.ParseConfig(connString)
	if err != nil {
		return nil, err
	}
	return t.ConnectConfig(ctx, connConfig)
}

// ConnectConfig connects to database or replay file with a configuration struct.
func (t *Transport) ConnectConfig(ctx context.Context, connConfig *pgx.ConnConfig) (conn Postgres, err error) {
	if t.Bypass {
		if t.Update {
			return nil, errors.New("cannot bypass when updating playback files")
		}
		if t.conn, err = pgx.ConnectConfig(ctx, connConfig); err != nil {
			return nil, fmt.Errorf("failed to connect to database: %w", err)
		}
		return t.conn, nil
	}
	t.script = &pgmock.Script{}
	if t.Update {
		return t.update(ctx, connConfig)
	}
	return t.playback(ctx)
}

func (t *Transport) playback(ctx context.Context) (conn Postgres, err error) {
	// Read the script to be replayed.
	data, err := os.ReadFile(t.Name)
	if err != nil {
		return nil, err
	}

	// Read file line-by-line, and try to create the replay script from it.
	for line, row := range bytes.Split(data, []byte("\n")) {
		var step pgmock.Step
		switch {
		case len(row) == 0 || bytes.HasPrefix(row, []byte("# ")):
			continue
		case bytes.HasPrefix(row, []byte("B ")):
			var msg pgproto3.BackendMessage
			if msg, err = t.getBackendMessage(row[2:]); err != nil {
				return nil, fmt.Errorf("line %d: %w", line, err)
			}
			step = pgmock.SendMessage(msg)
		case bytes.HasPrefix(row, []byte("F ")):
			var msg pgproto3.FrontendMessage
			if msg, err = t.getFrontendMessage(row[2:]); err != nil {
				return nil, fmt.Errorf("line %d: %w", line, err)
			}

			// TODO(henvic): Determine how to use ExpectMessage and ExpectAnyMessage in a better way.
			// Possibly read from an inline JSON param.
			if t.Strict {
				step = pgmock.ExpectMessage(msg)
				break
			}
			switch msg.(type) {
			case *pgproto3.StartupMessage, *pgproto3.Bind:
				step = pgmock.ExpectAnyMessage(msg)
			default:
				step = pgmock.ExpectMessage(msg)
			}
		default:
			return nil, fmt.Errorf("line %d: unexpected value (%q)", line, row)
		}
		t.script.Steps = append(t.script.Steps, step)
	}
	t.log(len(t.script.Steps), "steps loaded")
	if err := t.listenProxy(ctx); err != nil {
		return nil, err
	}
	go func() {
		defer close(t.proxyErrChan)
		conn, err := t.listener.Accept()
		if err != nil {
			t.proxyErrChan <- err
			return
		}

		t.log(("Running replay script"))
		backend := pgproto3.NewBackend(pgproto3.NewChunkReader(conn), conn)
		err = t.script.Run(backend)
		defer func() {
			// Ideally, it should set an error to upcoming messages to mark connection as terminated prematurely.
			if t.conn != nil {
				// Don't wait, close connection immediately.
				expiredCtx, cancel := context.WithCancel(context.Background())
				cancel()
				t.conn.Close(expiredCtx)
			}
		}()
		if err != nil {
			t.proxyErrChan <- err
			return
		}
		t.proxyErrChan <- nil
	}()

	addr := strings.Split(t.listener.Addr().String(), ":")
	host, port := addr[0], addr[1]
	connStr := fmt.Sprintf("sslmode=disable host=%s port=%s", host, port)
	t.log("Connecting to playback:", connStr)
	if t.conn, err = pgx.Connect(ctx, connStr); err != nil {
		err := fmt.Errorf("cannot playback: %w", err)
		if ec, ok := <-t.proxyErrChan; ok {
			err = fmt.Errorf("%v: %w", err, ec)
		}
		return nil, err
	}
	return t.conn, nil
}

// update uses most of the logic of github.com/jackc/pgmock/pgmockproxy
// to define a proxy without having to call an external process.
func (t *Transport) update(ctx context.Context, connConfig *pgx.ConnConfig) (conn Postgres, err error) {
	t.log("Updating playback file")
	if err := t.listenProxy(ctx); err != nil {
		return nil, err
	}
	if !strings.HasSuffix(t.Name, ".pgplayback") {
		return nil, errors.New("playback file must have .pgplayback suffix")
	}
	go func() {
		defer close(t.proxyErrChan)
		clientConn, err := t.listener.Accept()
		if err != nil {
			t.proxyErrChan <- err
			return
		}

		network := "tcp"
		if _, err := os.Stat(t.Remote); err == nil {
			network = "unix"
		}

		serverConn, err := net.Dial(network, t.Remote)
		if err != nil {
			t.proxyErrChan <- fmt.Errorf("cannot connect to database: %w", err)
			return
		}

		proxy := proxy.NewProxy(clientConn, serverConn)

		f, err := os.Create(t.Name)
		if err != nil {
			t.proxyErrChan <- fmt.Errorf("cannot create playback file: %w", err)
			return
		}

		var w io.Writer = f
		if t.Debug {
			w = io.MultiWriter(os.Stdout, f)
		}

		if err = proxy.Stream(w); err != nil {
			t.proxyErrChan <- fmt.Errorf("error running proxy: %w", err)
			return
		}
		t.proxyErrChan <- nil
	}()

	// TODO(henvic): Take a look at pgmock.AcceptUnauthenticatedConnRequestSteps() to check how credentials are used,
	// and consider redacting them by default.
	connConfig.LookupFunc = func(ctx context.Context, host string) (addrs []string, err error) {
		return []string{"127.0.0.1"}, nil
	}
	connConfig.DialFunc = func(ctx context.Context, network, addr string) (net.Conn, error) {
		var d net.Dialer
		return d.DialContext(ctx, "tcp", t.listener.Addr().String())
	}
	// NOTE(henvic): This approach doesn't work with TLS. Ideally, just tap the wire communication directly,
	// instead of creating a proxy.
	connConfig.TLSConfig = nil

	if t.conn, err = pgx.ConnectConfig(ctx, connConfig); err != nil {
		return nil, fmt.Errorf("cannot connect to database using proxy: %w", err)
	}
	return t.conn, nil
}

// listenProxy start a fake Postgres server on a fake TCP port accessible locally when used on playback mode
// and serves as proxy when using for updating the playback files.
// TODO: maybe update doc.
func (t *Transport) listenProxy(ctx context.Context) (err error) {
	var lc net.ListenConfig
	t.listener, err = lc.Listen(ctx, "tcp", "127.0.0.1:")
	if err != nil {
		return fmt.Errorf("cannot start replay proxy: %w", err)
	}
	t.proxyErrChan = make(chan error, 1)
	return nil
}

// Close database connection or replay session.
func (t *Transport) Close(ctx context.Context) (err error) {
	if !t.Bypass && !t.Update {
		return t.closePlayback(ctx)
	}
	if t.conn != nil {
		defer func() {
			ec := t.conn.Close(ctx)
			if err == nil {
				err = ec
			}
		}()
		select {
		case err := <-t.proxyErrChan:
			return err
		case <-ctx.Done():
			return ctx.Err()
		default:
			return nil
		}
	}
	return nil
}

func (t *Transport) closePlayback(ctx context.Context) (err error) {
	defer func() {
		if t.listener != nil {
			t.listener.Close()
		}
	}()
	defer func() {
		if t.conn != nil {
			// Don't wait, close connection immediately.
			expiredCtx, cancel := context.WithCancel(context.Background())
			cancel()
			t.conn.Close(expiredCtx)
		}
	}()

	select {
	case err := <-t.proxyErrChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	default:
		// TODO(henvic): check if we need to print something along the lines of
		// "playback closed before last step was executed" when the connection is prematurely terminated.
		return nil
	}
}

func (t *Transport) log(a ...interface{}) {
	if t.Debug {
		fmt.Println(a...)
	}
}
