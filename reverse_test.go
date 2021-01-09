package reverse_test

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"testing"

	"insite.feed/zxdev/reverse"
)

func TestReverse(t *testing.T) {

	var max, count = 3, 0

	var b bytes.Buffer
	//b.WriteString("\n") // empty
	for i := 0; i < max; i++ {
		b.WriteString(fmt.Sprintf("line-%x\n", i))
	}

	scanner := reverse.NewScanner(bytes.NewReader(b.Bytes()), b.Len(), nil)
	for scanner.Scan() {
		log.Println(count, scanner.Text(), scanner.Bytes(), scanner.Err())
		count++
	}

	// === RUN   TestReverse
	// 2021/01/08 22:22:27 0 line-2 [108 105 110 101 45 50] <nil>
	// 2021/01/08 22:22:27 1 line-1 [108 105 110 101 45 49] <nil>
	// 2021/01/08 22:22:27 2 line-0 [108 105 110 101 45 48] EOF
	// --- PASS: TestReverse (0.00s)

}

func TestReverseFile(t *testing.T) {

	var max, count = 3, 0
	var file = "sample.log"
	defer os.Remove(file)

	w, _ := os.Create(file)
	//w.WriteString("\n") // empty
	for i := 0; i < max; i++ {
		w.WriteString(fmt.Sprintf("line-%x\n", i))
	}
	w.Close()

	f, _ := os.Open(file)
	info, _ := f.Stat()
	defer f.Close()

	scanner := reverse.NewScanner(f, int(info.Size()), &reverse.Options{ChunkSize: 10, BufferSize: 100})
	for scanner.Scan() {
		log.Println(count, scanner.Text(), scanner.Bytes(), scanner.Err())
		count++
	}

	// 	=== RUN   TestReverseFile
	// 2021/01/08 22:22:27 0 line-2 [108 105 110 101 45 50] <nil>
	// 2021/01/08 22:22:27 1 line-1 [108 105 110 101 45 49] <nil>
	// 2021/01/08 22:22:27 2 line-0 [108 105 110 101 45 48] EOF
	// --- PASS: TestReverseFile (0.00s)

}
