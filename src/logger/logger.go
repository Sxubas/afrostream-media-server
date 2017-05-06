// Copyright (c) 2015
//      Sebastien Petit & Afrostream - www.afrostream.tv - spebsd@gmail.com.
//      All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions
// are met:
//
// 1. Redistributions of source code must retain the above copyright
//    notice, this list of conditions and the following disclaimer.
//
// 2. Redistributions in binary form must reproduce the above copyright
//    notice, this list of conditions and the following disclaimer in the
//    documentation and/or other materials provided with the distribution.
//
// 3. Neither the name of the copyright holder nor the names of its contributors
//    may be used to endorse or promote products derived from this software
//    without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
// FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
// DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS
// OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION)
// HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT
// LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY
// OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF
// SUCH DAMAGE.

package logger

import (
    "fmt"
    "log"
    "os"
)

const (
    F_Debug = 1 << iota
)

var (
    flag int

    prefixInfo string
    prefixDebug string
    prefixError string

    loggerInfo *log.Logger
    loggerError *log.Logger
)

func Message(format string, v ...interface{}) {
    fmt.Printf(format + "\n", v...)
}

func Info(format string, v...interface{}) {
    loggerInfo.Printf(prefixInfo + format, v...)
}

func Debug(format string, v ...interface{}) {
    if flag & F_Debug != 0 {
        fmt.Printf(prefixDebug + format + "\n", v...)
    }
}

func Error(format string, v...interface{}) {
    loggerError.Printf(prefixError + format, v...)
}

func Init(filename string, flags int) {
    flag = flags

    prefixInfo = "[ INFO ] "
    prefixDebug = "[ DEBUG ] "
    prefixError = "[ ERROR ] "

    loggerInfo = log.New(os.Stdout, "", log.LstdFlags)
    loggerError = log.New(os.Stderr, "", log.LstdFlags)

    if filename == "" {
        return
    }

    file, err := os.OpenFile(filename, os.O_WRONLY | os.O_CREATE | os.O_APPEND, 0644)
    if err != nil {
        Error("Couldn't create/open file %s : %s", filename, err)
    }

    loggerInfo.SetOutput(file)
    loggerError.SetOutput(file)

    defer file.Close()
}

