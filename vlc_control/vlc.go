package vlc

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

type Client struct {
	conn   net.Conn
	reader *bufio.Reader
}

// Connects to VLC's RC interface.
func NewClient(addr string) (*Client, error) {
	conn, err := net.DialTimeout("tcp", addr, 1*time.Second)
	if err != nil {
		return nil, err
	}

	c := &Client{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}

	// VLC sends a startup banner; read it but ignore
	conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	c.reader.ReadString('>') // RC prompt ends with "> "
	conn.SetReadDeadline(time.Time{})

	return c, nil
}

func (c *Client) send(cmd string) (string, error) {
	_, err := fmt.Fprintf(c.conn, "%s\n", cmd)
	if err != nil {
		return "", err
	}

	// Read until next prompt "> "
	var sb strings.Builder
	for {
		line, err := c.reader.ReadString('\n')
		if err != nil {
			return sb.String(), err
		}
		if strings.HasSuffix(line, "> ") {
			break
		}
		sb.WriteString(line)
	}
	return sb.String(), nil
}

// --- Playback Controls ---

func (c *Client) Play() error {
	_, err := c.send("play")
	return err
}

func (c *Client) Pause() error {
	_, err := c.send("pause")
	return err
}

func (c *Client) Stop() error {
	_, err := c.send("stop")
	return err
}

// --- Seeking ---

func (c *Client) SeekSeconds(sec int) error {
	_, err := c.send(fmt.Sprintf("seek %d", sec))
	return err
}

// Relative seek: +10 or -5
func (c *Client) SeekRelative(delta int) error {
	sign := ""
	if delta > 0 {
		sign = "+"
	}
	_, err := c.send(fmt.Sprintf("seek %s%d", sign, delta))
	return err
}

// --- Load file ---

func (c *Client) Load(path string) error {
	_, err := c.send(fmt.Sprintf("add %q", path))
	return err
}

// --- Query status ---

func (c *Client) Status() (string, error) {
	return c.send("status")
}

// --- Shutdown ---

func (c *Client) Close() error {
	return c.conn.Close()
}
