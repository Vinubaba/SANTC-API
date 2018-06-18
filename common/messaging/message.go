package messaging

import "time"

type Message struct {
	ID             string
	Data           []byte
	Attributes     map[string]string
	PublishTime    time.Time
	registeredAck  func() error
	registeredNack func() error
}

func (m Message) Ack() error {
	if m.registeredAck != nil {
		return m.registeredAck()
	}
	return nil
}

func (m Message) Nack() error {
	if m.registeredAck != nil {
		return m.registeredNack()
	}
	return nil
}

func (m *Message) RegisterAck(f func() error) {
	m.registeredAck = f
}

func (m *Message) RegisterNack(f func() error) {
	m.registeredNack = f
}
