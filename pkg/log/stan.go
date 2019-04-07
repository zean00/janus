package log

import (
	"fmt"

	stan "github.com/nats-io/go-nats-streaming"
	"github.com/sirupsen/logrus"
)

// NatsHook will emit logs to the subject provided
type NatsHook struct {
	conn          stan.Conn
	subject       string
	extraFields   map[string]interface{}
	dynamicFields map[string]func() interface{}
	Formatter     logrus.Formatter

	LogLevels []logrus.Level
}

// NewNatsHook will create a logrus hook that will automatically send
// new info into the channel
func NewNatsHook(conn stan.Conn, subject string) *NatsHook {
	hook := NatsHook{
		conn:          conn,
		subject:       subject,
		extraFields:   make(map[string]interface{}),
		dynamicFields: make(map[string]func() interface{}),
		Formatter:     &logrus.JSONFormatter{},
		LogLevels: []logrus.Level{
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
			logrus.WarnLevel,
			logrus.InfoLevel,
			logrus.DebugLevel,
		},
	}

	return &hook
}

// AddField will add a simple value each emission
func (hook *NatsHook) AddField(key string, value interface{}) *NatsHook {
	hook.extraFields[key] = value
	return hook
}

// AddDynamicField will call that method on each fire
func (hook *NatsHook) AddDynamicField(key string, generator func() interface{}) *NatsHook {
	hook.dynamicFields[key] = generator
	return hook
}

// Fire will use the connection and try to send the message to the right destination
func (hook *NatsHook) Fire(entry *logrus.Entry) error {
	if hook.conn.NatsConn().IsClosed() {
		return fmt.Errorf("Attempted to log on a closed connection")
	}

	// add in the new fields
	for k, v := range hook.extraFields {
		entry.Data[k] = v
	}

	for k, generator := range hook.dynamicFields {
		entry.Data[k] = generator()
	}

	bytes, err := hook.Formatter.Format(entry)
	if err != nil {
		return err
	}

	return hook.conn.Publish(hook.subject, bytes)
}

// Levels will describe what levels the NatsHook is associated with
func (hook *NatsHook) Levels() []logrus.Level {
	return hook.LogLevels
}
