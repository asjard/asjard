package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/asjard/asjard/pkg/security"
)

var (
	h bool   // show help
	f string // input file
	d bool   // decrypt
	o string // output
	t string // encrypt or decrypt text
	k string
	v string
	q bool
)

func usage() {
	fmt.Fprintf(os.Stderr, `asjard_cipher_aes version:
asjard_cipher_aes/1.0.0
Usage: asjard_cipher_aes [-hd] [-f input_file] [-o output_file] [-t text] <-k key> [-v iv]

Options:
`)
	flag.PrintDefaults()
}

func main() {
	flag.BoolVar(&h, "h", false, "this help")
	flag.StringVar(&f, "f", "", "encrypt or decrypt file")
	flag.BoolVar(&d, "d", false, "decrypt")
	flag.StringVar(&o, "o", "", "output file, default f's value + '_encrypted' or '_decrypted'")
	flag.StringVar(&t, "t", "", "encryp or decrypt text")
	flag.StringVar(&k, "k", "", "key, base64 encoded, length 16 or 24 or 32")
	flag.StringVar(&v, "v", "", "offset, base64 encoded, default k's value")
	flag.BoolVar(&q, "q", false, "quite output")
	flag.Usage = usage
	flag.Parse()
	if h {
		flag.Usage()
		os.Exit(0)
	}
	if k == "" || (t == "" && f == "") {
		fmt.Printf("k is must, t and f must have one\n\n")
		flag.Usage()
		os.Exit(1)
	}
	c, err := security.MustNewAESCipher(k, v)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	if t != "" {
		if !d {
			result, err := c.Encrypt(t, nil)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			if q {
				fmt.Print(result)
			} else {
				fmt.Println("encrypt text SUCCESS, base64 output:", result)
			}
		} else {
			result, err := c.Decrypt(t, nil)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			if q {
				fmt.Print(result)
			} else {
				fmt.Println("decrypt text SUCCESS, output:", result)
			}

		}
	}

	encryptedFileNamePrefix := "encrypted_" + security.AESCipherName + "_"
	decryptedFileNamePrefix := "decrypted_" + security.AESCipherName + "_"
	if f != "" {
		if !d {
			if o == "" {
				dir := filepath.Dir(f)
				fileName := filepath.Base(f)
				o = filepath.Join(dir, encryptedFileNamePrefix+fileName)
			}
			content, err := os.ReadFile(f)
			if err != nil {
				fmt.Println("read file", f, "fail", err.Error())
				os.Exit(1)
			}
			result, err := c.Encrypt(string(content), nil)
			if err != nil {
				fmt.Println("encrypt file", f, "fail", err.Error())
				os.Exit(1)
			}
			if err := os.WriteFile(o, []byte(result), 0400); err != nil {
				fmt.Println("write file", o, "fail", err.Error())
				os.Exit(1)
			}
			if !q {
				fmt.Println("encrypt file", f, "SUCCESS, output:", o)
			}
		} else {
			if o == "" {
				dir := filepath.Dir(f)
				fileName := filepath.Base(f)
				if strings.HasPrefix(fileName, encryptedFileNamePrefix) {
					o = filepath.Join(dir, strings.TrimPrefix(fileName, encryptedFileNamePrefix))
				} else {
					o = filepath.Join(dir, decryptedFileNamePrefix+fileName)
				}
			}
			content, err := os.ReadFile(f)
			if err != nil {
				fmt.Println("read file", f, "fail", err.Error())
				os.Exit(1)
			}
			result, err := c.Decrypt(string(content), nil)
			if err != nil {
				fmt.Println("decrypt file", f, "fail", err.Error())
				os.Exit(1)
			}
			if err := os.WriteFile(o, []byte(result), 0400); err != nil {
				fmt.Println("write file", o, "fail", err.Error())
				os.Exit(1)
			}
			if !q {
				fmt.Println("decrypt file", f, "SUCCESS, output:", o)
			}
		}
	}
}
