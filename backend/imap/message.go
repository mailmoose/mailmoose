package imapbackend

import (
	"bufio"
	"bytes"
	"fmt"
	"io"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/backend/backendutil"
	"github.com/emersion/go-message"
	"github.com/emersion/go-message/textproto"
	"github.com/mistralmail/mistralmail/backend/models"
)

type IMAPMessage struct {
	message     *models.Message
	messageRepo *models.MessageRepository
}

func NewIMAPMessageFromMessage(message *models.Message, messageRepo *models.MessageRepository) IMAPMessage {
	return IMAPMessage{
		message:     message,
		messageRepo: messageRepo,
	}
}

func (m *IMAPMessage) entity() (*message.Entity, error) {
	return message.Read(bytes.NewReader(m.message.Body))
}

func (m *IMAPMessage) headerAndBody() (textproto.Header, io.Reader, error) {

	// If the body isn't selected from the database, we need to get the message with its body
	if m.message.Body == nil {
		messageWithBody, err := m.messageRepo.GetMessageByID(m.message.ID)
		if err != nil {
			return textproto.Header{}, nil, fmt.Errorf("couldn't get message body: %w", err)
		}
		m.message.Body = messageWithBody.Body
	}

	body := bufio.NewReader(bytes.NewReader(m.message.Body))
	hdr, err := textproto.ReadHeader(body)
	return hdr, body, err
}

func (m *IMAPMessage) Fetch(seqNum uint32, items []imap.FetchItem) (*imap.Message, error) {

	fetched := imap.NewMessage(seqNum, items)
	for _, item := range items {
		switch item {
		case imap.FetchEnvelope:
			hdr, _, err := m.headerAndBody()
			if err != nil {
				return nil, fmt.Errorf("couldn't get header and body: %w", err)
			}
			fetched.Envelope, _ = backendutil.FetchEnvelope(hdr)
		case imap.FetchBody, imap.FetchBodyStructure:
			hdr, body, _ := m.headerAndBody()
			fetched.BodyStructure, _ = backendutil.FetchBodyStructure(hdr, body, item == imap.FetchBodyStructure)
		case imap.FetchFlags:
			fetched.Flags = m.message.Flags
		case imap.FetchInternalDate:
			fetched.InternalDate = m.message.Date
		case imap.FetchRFC822Size:
			fetched.Size = m.message.Size
		case imap.FetchUid:
			fetched.Uid = uint32(m.message.ID)
		default:
			section, err := imap.ParseBodySectionName(item)
			if err != nil {
				break
			}

			body := bufio.NewReader(bytes.NewReader(m.message.Body))
			hdr, err := textproto.ReadHeader(body)
			if err != nil {
				return nil, err
			}

			l, _ := backendutil.FetchBodySection(hdr, body, section)
			fetched.Body[section] = l
		}
	}

	return fetched, nil
}

func (m *IMAPMessage) Match(seqNum uint32, c *imap.SearchCriteria) (bool, error) {
	e, _ := m.entity()
	return backendutil.Match(e, seqNum, uint32(m.message.ID), m.message.Date, m.message.Flags, c)
}
