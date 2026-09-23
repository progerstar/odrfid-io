# odrfid-io

[Русский](#русский) · [English](#english)

## Русский

`odrfid-io` читает UID метки из последовательного порта считывателя ODRFID-MNE, вводит выбранный формат UID в активное поле и по умолчанию нажимает Enter. Программа принимает UID длиной от 4 до 10 байт.

### Запуск

Для сборки нужны Go 1.23.5 или новее и C-компилятор для RobotGo. Считыватель должен поддерживать команды `AT+H0`, `AT+F=HU*` и `AT+SCAN2`. Закройте другие программы, которые используют его порт.

Linux:

```sh
go build -o odrfid-io .
./odrfid-io -port /dev/ttyACM0
```

В Windows соберите программу командой `go build -o odrfid-io.exe .` и запустите `.\odrfid-io.exe -port COM3`. Порт, режим и нажатие Enter можно задать в `start.bat`. На macOS используйте команды для Linux с портом вида `/dev/cu.usbmodem...` и разрешите терминалу управление компьютером в настройках универсального доступа.

В Linux для сборки нужны заголовки X11/XTest, для работы — сеанс X11 и, при стандартном способе ввода, `xclip` или `xsel`. Когда появится сообщение `ожидание UID`, установите фокус в нужное поле и поднесите метку. Остановите программу сочетанием Ctrl+C.

### Форматы UID

По умолчанию режим `id` вводит двухзначную длину UID в байтах и десятичное значение всех его байтов в прямом порядке. Например, `7A403AB9` становится `042051029689`. Сохраняйте такой номер как текст, чтобы не потерять ведущий ноль.

| `-mode` | Вывод для `7A403AB9` | Значение |
| --- | --- | --- |
| `id` | `042051029689` | Полный UID с длиной; режим по умолчанию. |
| `hex` | `7A403AB9` | Полный UID в верхнем HEX-регистре. |
| `hex-reverse` | `B93A407A` | Байты UID в обратном порядке. |
| `decimal` | `2051029689` | Десятичное число, прямой порядок байтов. |
| `decimal-le` | `3107602554` | Десятичное число, обратный порядок байтов. |
| `wiegand26` | `058,16506` | Для UID из 4 байт: код объекта и номер карты. |
| `em4100` | — | Для UID из 5 байт; пример: `0102034050` → `003,16464`. |

`wiegand26` вводит два текстовых поля, а не сигнал по линиям Wiegand. Он отбрасывает четвёртый байт UID, а `em4100` — первые два: разные UID могут дать одинаковый номер. Для UID другой длины эти режимы пропускают метку. `decimal` и `decimal-le` также не сохраняют длину UID при ведущих или замыкающих нулевых байтах. Если нужно сохранить полный UID, используйте `id` или `hex`.

Для собственной строки укажите `-format` вместо `-mode`:

```sh
./odrfid-io -port /dev/ttyACM0 -format '{id},{uid},{decimal-le}'
```

Доступны поля `{id}`, `{uid}`, `{uid-reverse}`, `{decimal}`, `{decimal-le}`, `{length}`, `{wiegand26}` и `{em4100}`. Параметр `-sep` меняет разделитель внутри `wiegand26` и `em4100` (по умолчанию запятая). Если поле не подходит к длине UID, строка не вводится. Прежние режимы `default`, `conv1` и `conv2` работают как `hex`, `decimal-le` и `wiegand26`.

### Ввод и соединение

По умолчанию после каждого UID программа нажимает Enter. Чтобы оставить ввод в активном поле без Enter, добавьте `-enter=false`, например `./odrfid-io -port /dev/ttyACM0 -enter=false`. Флаг `-enter=true` включает нажатие обратно.

По умолчанию в Windows и macOS RobotGo вводит символы, а в Linux/X11 программа вставляет строку через буфер обмена. Вставка **заменяет содержимое текстового буфера обмена**. Если приложение не принимает вставку, укажите `-input=type`; тогда на Linux буквы могут зависеть от раскладки клавиатуры. Режим `id` состоит только из цифр. Нативный ввод через Wayland не поддерживается.

Программа открывает порт, отправляет `AT+H0`, `AT+F=HU*` и `AT+SCAN2`, затем принимает полные UID в верхнем HEX-регистре. Ответы на команды и события удаления метки она пропускает. При ошибке порта программа повторяет подключение через 2 секунды. После выхода клавиатура считывателя остаётся отключённой; включить её можно командой `AT+H1` через последовательный порт.

Лицензия: [MIT](LICENSE).

## English

`odrfid-io` reads tag UIDs from an ODRFID-MNE reader's serial port, types the selected UID format into the focused field, and presses Enter by default. It accepts UIDs from 4 to 10 bytes long.

### Run

Building requires Go 1.23.5 or newer and a C compiler for RobotGo. The reader must support `AT+H0`, `AT+F=HU*`, and `AT+SCAN2`. Close other programs using its port.

Linux:

```sh
go build -o odrfid-io .
./odrfid-io -port /dev/ttyACM0
```

On Windows, build with `go build -o odrfid-io.exe .` and run `.\odrfid-io.exe -port COM3`. You can set the port, mode, and Enter behavior in `start.bat`. On macOS, use the Linux commands with a port such as `/dev/cu.usbmodem...`, and grant your terminal Accessibility permission.

Linux builds need X11/XTest headers. Running requires an X11 session and, with the default input method, `xclip` or `xsel`. When the program reports `ожидание UID` (“waiting for UID”), focus the destination field and present a tag. Press Ctrl+C to stop.

### UID formats

By default, `id` types a two-digit UID length in bytes followed by the decimal value of all UID bytes in their original order. For example, `7A403AB9` becomes `042051029689`. Store this number as text to preserve its leading zero.

| `-mode` | Output for `7A403AB9` | Meaning |
| --- | --- | --- |
| `id` | `042051029689` | Complete UID with its length; default. |
| `hex` | `7A403AB9` | Full uppercase hexadecimal UID. |
| `hex-reverse` | `B93A407A` | UID bytes in reverse order. |
| `decimal` | `2051029689` | Decimal number, original byte order. |
| `decimal-le` | `3107602554` | Decimal number, reversed byte order. |
| `wiegand26` | `058,16506` | For a 4-byte UID: facility code and card number. |
| `em4100` | — | For a 5-byte UID; example: `0102034050` → `003,16464`. |

`wiegand26` types two text fields; it does not send a signal over Wiegand wires. It drops the fourth UID byte, while `em4100` drops the first two, so different UIDs can produce the same number. These modes skip tags with other UID lengths. `decimal` and `decimal-le` can also lose UID length when leading or trailing zero bytes are present. Use `id` or `hex` to keep the complete UID.

Use `-format` instead of `-mode` to compose a line:

```sh
./odrfid-io -port /dev/ttyACM0 -format '{id},{uid},{decimal-le}'
```

Available fields: `{id}`, `{uid}`, `{uid-reverse}`, `{decimal}`, `{decimal-le}`, `{length}`, `{wiegand26}`, and `{em4100}`. `-sep` changes the separator within `wiegand26` and `em4100` (a comma by default). If a field does not apply to the UID length, nothing is typed. Legacy modes `default`, `conv1`, and `conv2` work as `hex`, `decimal-le`, and `wiegand26`.

### Input and connection

By default, the program presses Enter after each UID. Add `-enter=false` to leave the text in the focused field without Enter, for example `./odrfid-io -port /dev/ttyACM0 -enter=false`. Use `-enter=true` to enable it again.

By default, RobotGo types characters on Windows and macOS; on Linux/X11, the program pastes the line through the clipboard. Pasting **replaces the current text clipboard contents**. If the destination blocks paste, use `-input=type`; on Linux, letters may then depend on the keyboard layout. The `id` mode contains digits only. Native Wayland input is not supported.

The program opens the port, sends `AT+H0`, `AT+F=HU*`, and `AT+SCAN2`, then accepts complete uppercase hexadecimal UIDs. It skips command replies and tag removal events. After a port error, it reconnects in 2 seconds. The reader's keyboard remains disabled when the program exits; send `AT+H1` over the serial port to re-enable it.

License: [MIT](LICENSE).
