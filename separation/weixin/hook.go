package weixin

import (
	"bufio"
	"bytes"
	jsoniter "github.com/json-iterator/go"
	"net/http"
	"strings"
)

const (
	textLimit     = 2040 // 文本内容，最长不超过2048个字节，必须是utf8编码
	markdownLimit = 4090 // markdown内容，最长不超过4096个字节，必须是utf8编码
	baseUrl       = "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key="
)

type WebhookMsg struct {
	MsgType  string       `json:"msgtype"`
	Text     *TextMsg     `json:"text,omitempty"`
	Markdown *MarkdownMsg `json:"markdown,omitempty"`
	Id       string       `json:"-"`
}

type TextMsg struct {
	Content             string   `json:"content"`
	MentionedList       []string `json:"mentioned_list"`
	MentionedMobileList []string `json:"mentioned_mobile_list"`
}

type MarkdownMsg struct {
	Content string `json:"content"`
}

func New() *WebhookMsg {
	return &WebhookMsg{}
}
func (m *WebhookMsg) ConvertAndSend(payload []byte, id string) error {
	m.Id = id
	err := m.unmarshal(payload)
	if err != nil {
		return m.send(payload)
	}
	switch m.MsgType {
	case "text":
		return m.textHandle(payload)
	case "markdown":
		return m.markdownHandle(payload)
	default:
		return m.send(payload)
	}
}

func (m *WebhookMsg) textHandle(payload []byte) error {
	content := m.Text.Content
	if len(content) <= textLimit {
		return m.send(payload)
	}
	reader := bufio.NewScanner(strings.NewReader(content))
	var (
		buf         bytes.Buffer
		notLastLine = reader.Scan()
	)
	for notLastLine {
		line := reader.Text()
		if buf.Len()+len(line) <= textLimit {
			buf.WriteString(line)
			if notLastLine {
				buf.WriteString("\n")
			}
		} else {
			if err := m.textSendLoop(buf.String()); err != nil {
				return err
			}
			if len(line) > markdownLimit {
				if err := m.textSendLoop(line); err != nil {
					return err
				}
			} else {
				buf.WriteString(line)
				buf.WriteString("\n")
			}
		}
		notLastLine = reader.Scan()
	}
	m.Text.Content = buf.String()
	return m.marshalAndSend()
}

func (m *WebhookMsg) markdownHandle(payload []byte) error {
	content := m.Markdown.Content
	if len(content) <= markdownLimit {
		return m.send(payload)
	}
	reader := bufio.NewScanner(strings.NewReader(content))
	var (
		buf         bytes.Buffer
		notLastLine = reader.Scan()
	)
	for notLastLine {
		line := reader.Text()
		if buf.Len()+len(line) <= markdownLimit {
			buf.WriteString(line)
			if notLastLine {
				buf.WriteString("\n")
			}
		} else {
			if err := m.markdownSendLoop(buf.String()); err != nil {
				return err
			}
			buf.Reset()
			if len(line) > markdownLimit {
				if err := m.markdownSendLoop(line); err != nil {
					return err
				}
			} else {
				buf.WriteString(line)
				buf.WriteString("\n")
			}
		}
		notLastLine = reader.Scan()
	}
	m.Markdown.Content = buf.String()
	return m.marshalAndSend()
}
func (m *WebhookMsg) textSendLoop(content string) error {
	if len(content) == 0 {
		return nil
	}
	if len(content) <= textLimit {
		m.Text.Content = content
		return m.marshalAndSend()
	}
	maxIndex := len(content) - 1
	n := maxIndex / textLimit
	for i := 0; i < n; i++ {
		m.Text.Content = content[i*textLimit : (i+1)*textLimit]
		if err := m.marshalAndSend(); err != nil {
			return err
		}
	}
	m.Text.Content = content[n*textLimit : maxIndex]
	return m.marshalAndSend()
}

func (m *WebhookMsg) markdownSendLoop(content string) error {
	if len(content) == 0 {
		return nil
	}
	if len(content) <= markdownLimit {
		m.Markdown.Content = content
		return m.marshalAndSend()
	}
	maxIndex := len(content) - 1
	n := maxIndex / markdownLimit
	for i := 0; i < n; i++ {
		m.Markdown.Content = content[i*markdownLimit : (i+1)*markdownLimit]
		if err := m.marshalAndSend(); err != nil {
			return err
		}
	}
	m.Markdown.Content = content[n*markdownLimit : maxIndex]
	return m.marshalAndSend()
}
func (m *WebhookMsg) unmarshal(payload []byte) error {
	return jsoniter.Unmarshal(payload, &m)
}

func (m *WebhookMsg) marshalAndSend() error {
	switch m.MsgType {
	case "text":
		if len(m.Text.Content) == 0 {
			return nil
		}
	case "markdown":
		if len(m.Markdown.Content) == 0 {
			return nil
		}
	}
	payload, err := jsoniter.Marshal(m)
	if err != nil {
		return err
	}
	return m.send(payload)
}

func (m *WebhookMsg) send(payload []byte) error {
	_, err := http.Post(baseUrl+m.Id, "application/json", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	return nil
}
