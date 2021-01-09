# reverse

A reverse (LIFO) line scanner to process a source line-by-line from tail-to-head that will accept any souce that can support an io.ReaderAt interface (eg. bytes.Buffer, io.File, etc) 

This reverse.Scanner operates similarly to a bufio.Scanner, however since this reads from tail-to-head you MUST tell it the size (or start position) for the source for proper retrograde line seek operation. 

```golang

	scanner := reverse.NewScanner(r, size, nil)
	for scanner.Scan() {
		log.Println(scanner.Text(), scanner.Bytes(), scanner.Err())
    }
    
    // 2021/01/08 21:58:08 0 line-2 [108 105 110 101 45 50] <nil>
    // 2021/01/08 21:58:08 1 line-1 [108 105 110 101 45 49] <nil>
    // 2021/01/08 21:58:08 2 line-0 [108 105 110 101 45 48] EOF

```

## Considerations

1) Passing a ```size``` of zero or less to reverse.NewScanner will result in a nil *Scanner being returned.

2) When nil is passed for ```opt``` the default ChunkSize, BufferSize, and IgnoreEmptyLine settsion will be used to configure reverse.NewScanner

* See test suite for sample use case examples.