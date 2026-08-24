package alarm

import (
	"fmt"
	"os"

	"cleanroomorcontrol/internal/model"
)

type LogNotifier struct {
	out *os.File
}

func NewLogNotifier() *LogNotifier {
	return &LogNotifier{out: os.Stdout}
}

func (n *LogNotifier) Notify(record model.AlarmRecord) error {
	_, err := fmt.Fprintf(n.out, "alarm %s level=%s room=%s active=%v message=%s\n",
		record.ID, record.Level, record.Room, record.Active, record.Message)
	return err
}
