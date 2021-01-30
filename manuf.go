package manuf

import (
	"bufio"
	"io"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/kardianos/osext"
)

const hexDigit = "0123456789ABCDEF"

var dGlobal map[int]interface{} = nil
var m sync.Mutex

// fileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func b2uint64(sList []string) uint64 {
	var t uint64
	for i, b := range sList {
		l := strings.Index(hexDigit, string(b[0]))
		r := strings.Index(hexDigit, string(b[1]))
		t += uint64((l<<4)+r) << uint8((6-i-1)*8)
	}

	return t
}

func initMapInternal() (map[int]interface{}, error) {
	ret := make(map[int]interface{})

	parse := func(mac, comment string) {
		g := strings.Split(mac, "/")
		m := strings.Split(g[0], ":")
		var b int
		if len(g) != 2 {
			b = 48 - len(m)*8
		} else {
			b, _ = strconv.Atoi(g[1])
		}
		if _, ok := ret[b]; !ok {
			ret[b] = make(map[uint64]string)

		}
		ret[b].(map[uint64]string)[b2uint64(m)] = comment
	}

	readLine := func(fileName string, handler func(string)) error {
		f, err := os.Open(fileName)
		defer f.Close()
		if err != nil {
			return err
		}
		buf := bufio.NewReader(f)
		for {
			line, err := buf.ReadString('\n')
			line = strings.TrimSpace(line)
			handler(line)
			if err != nil {
				if err == io.EOF {
					return nil
				}
				return err
			}
		}
	}

	f := "manuf"

	if !fileExists(f) {
		_, file, _, _ := runtime.Caller(0)
		f = path.Join(path.Dir(file), "manuf")
	}

	if !fileExists(f) {
		if exedir, err := osext.ExecutableFolder(); err == nil {
			for _, appender := range []string{
				"/manuf",
				"/common/manuf",
				"/../manuf",
				"/../common/manuf",
			} {
				f = exedir + appender

				if fileExists(f) {
					break
				}
			}
		}
	}

	err := readLine(f, func(s string) {
		l := strings.Split(s, "\t")
		if len(l) > 2 {
			parse(l[0], l[2])
		}
	})

	return ret, err
}

func initMap() error {
	if dGlobal != nil {
		return nil
	}

	m.Lock()
	defer m.Unlock()

	if dGlobal == nil {
		if d, e := initMapInternal(); e == nil {
			dGlobal = d
		} else {
			return e
		}
	}

	return nil
}

func Search(mac string) string {
	if err := initMap(); err != nil {
		return "failed"
	}

	mac = strings.ToUpper(mac)
	mac = strings.Replace(mac, "-", ":", -1)
	mac = strings.Replace(mac, " ", "", -1)

	if len(mac) != 17 {
		return ""
	}

	for i := 0; i < len(mac); i++ {
		if i%3 == 2 {
			if mac[i] != ':' {
				return ""
			}
		} else {
			if (mac[i] >= 'A' && mac[i] <= 'F') || (mac[i] >= '0' && mac[i] <= '9') {
				continue
			}

			return ""
		}
	}

	s := strings.Split(mac, ":")

	if len(s) != 6 {
		return ""
	}

	bint := b2uint64(s)
	for b := range dGlobal {
		k := 48 - b
		bint = (bint >> uint8(k)) << uint8(k)
		if _, ok := dGlobal[b].(map[uint64]string)[bint]; ok {
			return dGlobal[b].(map[uint64]string)[bint]
		}
	}
	return ""
}
