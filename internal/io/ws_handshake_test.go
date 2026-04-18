package io_test

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	slyio "github.com/slycrel/slycrel/internal/io"
	"golang.org/x/net/websocket"
)

// dialWS connects to a httptest.Server WebSocket endpoint.
func dialWS(t *testing.T, ts *httptest.Server) *websocket.Conn {
	t.Helper()
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/"
	ws, err := websocket.Dial(wsURL, "", "http://localhost/")
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	return ws
}

// sendHello marshals and sends a ClientMsg hello.
func sendHello(t *testing.T, ws *websocket.Conn, username, password string, version int) {
	t.Helper()
	msg := slyio.ClientMsg{
		Version:  version,
		Type:     slyio.MsgTypeHello,
		Username: username,
		Password: password,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal hello: %v", err)
	}
	if _, err := ws.Write(data); err != nil {
		t.Fatalf("write hello: %v", err)
	}
}

// recvMsg reads one ServerMsg from the WebSocket connection.
func recvMsg(t *testing.T, ws *websocket.Conn) slyio.ServerMsg {
	t.Helper()
	var raw []byte
	if err := websocket.Message.Receive(ws, &raw); err != nil {
		t.Fatalf("recv: %v", err)
	}
	var msg slyio.ServerMsg
	if err := json.Unmarshal(raw, &msg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return msg
}

// acceptAll is an Authenticator that accepts any credentials.
func acceptAll(_, _ string) error { return nil }

// rejectAll is an Authenticator that always rejects.
func rejectAll(_, _ string) error { return fmt.Errorf("authentication failed") }

// noopSession is a session handler that returns immediately.
func noopSession(_ *slyio.WSSession, _ string) {}

// TestWSHandshake_AuthOK verifies the full happy-path handshake sequence:
//   client → { type:"hello", v:1, username, password }
//   server → { type:"hello", v:1, seq:1 }
func TestWSHandshake_AuthOK(t *testing.T) {
	ts := httptest.NewServer(slyio.WSUpgrader(t.TempDir(), acceptAll, noopSession))
	defer ts.Close()

	ws := dialWS(t, ts)
	defer ws.Close()

	sendHello(t, ws, "hero", "secret", slyio.WSProtocolVersion)

	ack := recvMsg(t, ws)
	if ack.Type != slyio.MsgTypeHello {
		t.Errorf("expected hello ack, got type=%q err=%q", ack.Type, ack.ErrMsg)
	}
	if ack.Version != slyio.WSProtocolVersion {
		t.Errorf("ack version = %d, want %d", ack.Version, slyio.WSProtocolVersion)
	}
}

// TestWSHandshake_AuthFail verifies that wrong credentials produce an error message.
func TestWSHandshake_AuthFail(t *testing.T) {
	ts := httptest.NewServer(slyio.WSUpgrader(t.TempDir(), rejectAll, noopSession))
	defer ts.Close()

	ws := dialWS(t, ts)
	defer ws.Close()

	sendHello(t, ws, "hero", "wrongpass", slyio.WSProtocolVersion)

	msg := recvMsg(t, ws)
	if msg.Type != slyio.MsgTypeError {
		t.Errorf("expected error, got type=%q", msg.Type)
	}
	if msg.ErrMsg == "" {
		t.Error("expected non-empty error message")
	}
}

// TestWSHandshake_VersionMismatch verifies that a client with a wrong protocol version
// receives an error and the connection is closed.
func TestWSHandshake_VersionMismatch(t *testing.T) {
	ts := httptest.NewServer(slyio.WSUpgrader(t.TempDir(), acceptAll, noopSession))
	defer ts.Close()

	ws := dialWS(t, ts)
	defer ws.Close()

	sendHello(t, ws, "hero", "secret", 999)

	msg := recvMsg(t, ws)
	if msg.Type != slyio.MsgTypeError {
		t.Errorf("expected error, got type=%q", msg.Type)
	}
}

// TestWSHandshake_MissingUsername verifies that a hello without a username is rejected.
func TestWSHandshake_MissingUsername(t *testing.T) {
	ts := httptest.NewServer(slyio.WSUpgrader(t.TempDir(), acceptAll, noopSession))
	defer ts.Close()

	ws := dialWS(t, ts)
	defer ws.Close()

	sendHello(t, ws, "", "secret", slyio.WSProtocolVersion)

	msg := recvMsg(t, ws)
	if msg.Type != slyio.MsgTypeError {
		t.Errorf("expected error, got type=%q", msg.Type)
	}
}

// TestWSHandshake_SessionReceivesUsername verifies the session handler gets the
// correct username after a successful handshake.
func TestWSHandshake_SessionReceivesUsername(t *testing.T) {
	var gotUsername string
	done := make(chan struct{})

	handler := func(sess *slyio.WSSession, username string) {
		gotUsername = username
		close(done)
	}

	ts := httptest.NewServer(slyio.WSUpgrader(t.TempDir(), acceptAll, handler))
	defer ts.Close()

	ws := dialWS(t, ts)
	defer ws.Close()

	sendHello(t, ws, "aragorn", "strider", slyio.WSProtocolVersion)
	recvMsg(t, ws) // consume hello ack

	<-done
	if gotUsername != "aragorn" {
		t.Errorf("session got username %q, want %q", gotUsername, "aragorn")
	}
}
