package modem

// import (
// 	"github.com/RickyGrassmuck/modem_logger/modem/config"
// 	"github.com/RickyGrassmuck/modem_logger/modem/mb8611"
// 	"net/http"
// )

// type Modem interface {
// 	New(string, *http.Client, any) (*config.Config, error)
// 	GetLogs() (*LogData, error)
// }

// type LogData struct {
// 	Parsed string
// 	Raw    string
// }

// var modems map[string]func(any) (*config.Config, error)

// func init() {
// 	modems = map[string]func(any){
// 		"mb8611": mb8611.New,
// 	}
// }

// func NewModem(modemModel string, opts ...any) (Modem, error) {
// 	modem, ok := modems[modemModel](opts...)
// 	if !ok {
// 		return nil, fmt.Errorf("modem %s not found", modemModel)
// 	}
// 	return modem, nil
// }

// func (c *Config) GetLogs() (*LogData, error) {
// 	return c
// }

// func Authenticate() (bool, error) {
// 	return nil
// }
