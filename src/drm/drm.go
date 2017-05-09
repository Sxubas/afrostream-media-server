package main

import (
    "io"
    "fmt"
    "crypto/rand"
    "crypto/cipher"
    "crypto/aes"
)

func keygen(size uint) (b []byte, err error) {
    b = make([]byte, size)
    _, err = rand.Read(b)
    return
}

func encrypt(key []byte, iv [byte]data []byte) (crypted []byte, err error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }
    iv := []byte("AESkey0123456789")
    stream := cipher.NewCTR(block, iv)
    crypted = make([]byte, len(data))
    stream.XORKeyStream(crypted, data)
    return
}


func main() {
    data := []byte("")

    key, err := keygen(16)
    if err != nil {
        panic(err.Error())
    }

    crypted, err := encrypt(key, data)
    if err != nil {
        panic(err.Error())
    }

    decrypted, err := encrypt(key, crypted)
    if err != nil {
        panic(err.Error())
    }

    fmt.Printf("%x\n", key)
    fmt.Printf("%d : ''%s''\n", len(data), data)
    fmt.Printf("%d : ''%x''\n", len(crypted), crypted)
    fmt.Printf("%d : ''%s''\n", len(decrypted), decrypted)

    return
}

