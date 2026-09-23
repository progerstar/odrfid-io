package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"runtime"
	"strings"
	"time"

	"github.com/go-vgo/robotgo"
	"github.com/go-vgo/robotgo/clipboard"
	"github.com/tarm/serial"

	"odrfid-io/internal/conversion"
)

const reconnectDelay = 2 * time.Second

func main() {
	portName := flag.String("port", "COM3", "Port name /dev/ttyACM0, COM3, etc.")
	mode := flag.String("mode", "id", "Output mode: id, hex, hex-reverse, decimal, decimal-le, wiegand26, em4100")
	template := flag.String("format", "", "Output template; overrides -mode, e.g. '{id},{uid}'")
	sep := flag.String("sep", ",", "Separator for wiegand26 and em4100")
	input := flag.String("input", "auto", "Input method: auto, type, paste")
	pressEnter := flag.Bool("enter", true, "Press Enter after typing each UID")
	flag.Parse()

	formatter, err := conversion.NewFormatter(*mode, *template, *sep)
	if err != nil {
		log.Fatal(err)
	}
	if *input != "auto" && *input != "type" && *input != "paste" {
		log.Fatalf("неизвестный способ ввода %q", *input)
	}
	if (*input == "paste" || *input == "auto" && runtime.GOOS == "linux") && clipboard.Unsupported {
		log.Fatal("для вставки через буфер обмена нужен xclip или xsel; можно использовать -input=type")
	}

	config := &serial.Config{Name: *portName, Baud: 9600}
	for {
		if err := runPort(config, formatter, *input, *pressEnter); err != nil {
			log.Printf("порт %s: %v; повтор через %s", *portName, err, reconnectDelay)
		}
		time.Sleep(reconnectDelay)
	}
}

func runPort(config *serial.Config, formatter *conversion.Formatter, input string, pressEnter bool) error {
	port, err := serial.OpenPort(config)
	if err != nil {
		return fmt.Errorf("открытие: %w", err)
	}
	defer port.Close()

	if err := port.Flush(); err != nil {
		return fmt.Errorf("очистка буферов: %w", err)
	}
	// SCAN2 выдаёт форматированную строку при появлении метки; HU* задаёт полный HEX UID.
	scanner := bufio.NewScanner(port)
	for _, command := range []string{"AT+H0\r", "AT+F=HU*\r", "AT+SCAN2\r"} {
		if err := sendCommand(port, scanner, command); err != nil {
			return err
		}
	}

	log.Printf("порт %s открыт, ожидание UID", config.Name)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || line == "OK" || line == "ERROR" || strings.HasPrefix(line, "SCAN:-") {
			continue
		}
		uidText := strings.TrimPrefix(line, "SCAN:+")
		uid, err := conversion.ParseUID(uidText)
		if err != nil {
			log.Printf("пропущена строка от считывателя: %q", line)
			continue
		}
		output, err := formatter.Format(uid)
		if err != nil {
			log.Printf("преобразование UID %q: %v", uidText, err)
			continue
		}
		if err := typeOutput(output, input); err != nil {
			log.Printf("ввод UID %q: %v", uidText, err)
			continue
		}
		if pressEnter {
			if err := robotgo.KeyTap("enter"); err != nil {
				log.Printf("нажатие Enter после UID %q: %v", uidText, err)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("чтение: %w", err)
	}
	return io.EOF
}

func typeOutput(output, method string) error {
	if method == "auto" {
		if runtime.GOOS == "linux" {
			method = "paste"
		} else {
			method = "type"
		}
	}
	if method == "paste" {
		return robotgo.PasteStr(output)
	}
	robotgo.TypeStr(output)
	return nil
}

func sendCommand(port *serial.Port, scanner *bufio.Scanner, command string) error {
	n, err := port.Write([]byte(command))
	if err != nil {
		return fmt.Errorf("отправка %q: %w", command, err)
	}
	if n != len(command) {
		return fmt.Errorf("отправка %q: записано %d из %d байт", command, n, len(command))
	}

	for scanner.Scan() {
		switch line := scanner.Text(); line {
		case "":
			continue
		case "OK":
			return nil
		case "ERROR":
			return fmt.Errorf("считыватель отклонил %q", command)
		default:
			log.Printf("ответ до подтверждения %q: %q", command, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("ожидание ответа %q: %w", command, err)
	}
	return fmt.Errorf("ожидание ответа %q: %w", command, io.EOF)
}
