package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"fmt"
	"github.com/kbinani/screenshot"
	"golang.org/x/image/draw"
	"golang.org/x/mod/sumdb/dirhash"
	"image"
	"image/jpeg"
	"io"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const MockLength = 8
const MockChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// Key Version-specific and used in production binary
const Key = ""

// File used in source version and can be specific to user,
//const File = "key"

// Hashes Used in production binary
var Hashes = [][]string{
	{"cpu-z_2.03-en", "8dcd46e33f6f7be6c805bf701b7d557b0c92ec5a9e67381d82c909261d4467db1bcceaec4fbdf168a2f809b52a6698652127f962ffa376cd924aaf62fd6b8fd62f90cfc224fef174fe7dc29de22ca4966854ef9a9700e41665fc7b42f0618494"},
	{"", ""},
	{"PYPrime.2.0.Windows\\amd64", "6915656ddb470cb9dbfa4155804c6c6c5d5609458d8c958f0edf6c72da13fdd12a14bbfc6a07a6fb3a8d74959ba41975ca388b259a6c658a9178041a3bc1de71f60dd9f2fcbd495674dfc1555effb710eb081fc7d4cae5fa58c438ab50405081f4c788ff44eb904f041866a7aaf6d7bb6449b3acfd2581f0b50fda79b3898b34613e0d63b54ed995273eda446eb09e51066e486f1e72b94f1c338a83dca3a021c51c64dfb7c445ecf0001f69c27e13299ddcfba0780efa72b866a7487b7491c7"},
}
var Executables = []string{
	"cpuz_x64.exe",
	"",
	"PYPrime.exe",
}
var Configs = [][]string{
	{"cpuz.ini"},
}

func main() {
	if len(os.Args) > 1 {
		r, err := os.ReadFile(os.Args[1])
		if err == nil {
			c, err2 := strconv.Atoi(strings.Split(os.Args[1], "-")[0])
			if err2 == nil {
				r := xor(r, []byte(Key))
				i := int(r[0])
				if len(Executables) > i && i >= 0 {
					println(Executables[i])
					println(string(r[1 : 1+c]))
					if len(os.Args) > 2 {
						_ = os.WriteFile(os.Args[2], r[1+c:], 0666)
					}
				}
			}
		}
		return
	}

	c := exec.Command("cmd", "/c", "cls")
	c.Stdout = os.Stdout
	_ = c.Run()

	println("previousmark [-all] [resultfile] [output.jpg]  (v3 for Windows x64)")
	println("Press 1 for  CPU-Z      : memory frequency & CAS latency validation")
	println("Press 2 for  Clambench  : latency benchmark")
	print("Press 3 for  PYPrime    : general memory performance benchmark  >>> ")

	var option = ""
	var result = ""
	switch _, _ = fmt.Scanln(&option); option {
	case "1":
		_ = os.Remove(".\\" + Hashes[0][0] + "\\" + Configs[0][0])
		isGenuine(Hashes[0][0], Hashes[0][1]) // Used in production binary
		result = cpuz(".\\" + Hashes[0][0])
		launchForScreenshot(".\\"+Hashes[0][0]+"\\"+Executables[0], 3)
		encode(result, 0)
		println(result)
		return
	case "2":
		isGenuine(Hashes[1][0], Hashes[1][1])
		return
	case "3":
		isGenuine(Hashes[2][0], Hashes[2][1])
		result = pyprime(".\\" + Hashes[2][0])
		encode(result, 2)
		return
	default:
		return
	}
}

func cpuz(path string) string {
	freqSquared := 0.0
	cas := 0.0

	m := mockFilename()
	_ = exec.Command(path+"\\"+Executables[0], "-txt="+m).Run()
	defer func() { _ = os.Remove(path + "\\" + m + ".txt") }()

	f, err := os.Open(path + "\\" + m + ".txt")
	if err == nil {
		defer func() { _ = f.Close() }()
		newScanner := bufio.NewScanner(f)
		for newScanner.Scan() {
			line := strings.Replace(newScanner.Text(), "\t", " ", -1)
			if strings.Contains(line, "Memory Frequency") {
				s := strings.Trim(strings.SplitAfterN(line, "Memory Frequency", -1)[1], " ")
				freqSquared, err = strconv.ParseFloat(strings.Split(s, " ")[0], 64)
				if err != nil {
					panic("Malformed CPU-Z report")
				}
				freqSquared = math.Pow(freqSquared, 2)
			} else if strings.Contains(line, "CAS# latency (CL)") {
				s := strings.Trim(strings.SplitAfterN(line, "CAS# latency (CL)", -1)[1], " ")
				cas, err = strconv.ParseFloat(strings.Split(s, " ")[0], 64)
				if err != nil {
					panic("Malformed CPU-Z report")
				}
			}
		}
	}

	return fmt.Sprintf("%.3f", freqSquared/cas)
}

func clam() {

}

func pyprime(path string) string {
	_ = os.Chdir(path)
	defer func() { _ = os.Chdir(".\\..\\..") }()

	cmd := exec.Command("cmd")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    false,
		CmdLine:       fmt.Sprintf(` start "" /WAIT /HIGH /affinity 2 /B cmd /c .\%s 2B`, Executables[2]),
		CreationFlags: 0,
	}

	o, _ := cmd.Output()
	oo := string(o)
	if strings.Contains(oo, "10.0 MHz") && strings.Contains(oo, " VALID") {
		for _, s := range strings.Split(oo, "\n") {
			println(s)
			if strings.Contains(s, "Computation time :") {
				break
			}
		}
		u := strings.Split(
			strings.Trim(
				strings.SplitAfterN(
					oo, "Computation time :", -1)[1],
				" "),
			" ")[0]
		return u
	} else {
		panic("Invalid PYPrime run")
	}
}

func isGenuine(k string, v string) {
	var h = ""
	f, _ := dirhash.DirFiles(k, k)
	for _, ff := range f {
		fff, err := os.Open(ff)
		if err == nil {
			hh := sha256.New()
			if _, err := io.Copy(hh, fff); err == nil {
				h += fmt.Sprintf("%x", hh.Sum(nil))
			}
			_ = fff.Close()
		}
	}
	//println(h)
	if h != v {
		panic("Corrupt or missing benchmark files")
	}
}

func mockFilename() string {
	rand.Seed(time.Now().UnixNano())
	o := make([]byte, MockLength)
	for i := range o {
		o[i] = MockChars[rand.Intn(len(MockChars))]
	}
	return string(o)
}

func launchForScreenshot(path string, count int) {
	for i := 1; i <= count; i++ {
		go func() { _ = exec.Command(path).Run() }()
		print("Launched benchmark window, " + strconv.Itoa(count-i) + " left")
		_, _ = fmt.Scanln()
	}
}

func resize(src image.Image, dstSize image.Point) *image.RGBA {
	srcRect := src.Bounds()
	dstRect := image.Rectangle{
		Min: image.Point{},
		Max: dstSize,
	}
	dst := image.NewRGBA(dstRect)
	draw.CatmullRom.Scale(dst, dstRect, src, srcRect, draw.Over, nil)
	return dst
}

func encode(result string, i int) {
	out := mockFilename()

	r := []byte(result)
	x := append([]byte{byte(i)}, r...)
	//y, _ := os.ReadFile(FILE)
	y := []byte(Key)

	test, _ := screenshot.CaptureDisplay(0)
	test = resize(draw.Image(test), image.Point{X: 1280, Y: 720})
	buf := new(bytes.Buffer)
	_ = jpeg.Encode(buf, test, nil)
	x = xor(append(x, buf.Bytes()...), y)
	out = strconv.Itoa(len(r)) + "-" + out
	_ = os.WriteFile(out, x, 0666)

	//println(result)
}

func xor(x []byte, y []byte) []byte {
	yy := y
	for i := math.Ceil(float64(len(x) / len(y))); i > 0; i-- {
		yy = append(yy, y...)
	}
	for j := range x {
		x[j] = x[j] ^ yy[j]
	}
	return x
}
