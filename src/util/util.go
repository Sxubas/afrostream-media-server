package util

import (
    "errors"
    "path"
    "strings"
)

func SplitFilename(filename string) (string, string) {
    ext := path.Ext(filename)
    return filename[:len(filename) - len(ext)], ext
}

func Basename(filename string) string {
    return filename[:len(filename)-len(path.Ext(filename))]
}

func ParseBasename(filename string) (string, string, string, string, error) {
    splitted := strings.Split(filename, "_")

    l := len(splitted)
    if l < 4 {
        return "", "", "", "", errors.New("Invalid filename format")
    }

    return strings.Join(splitted[:len(splitted) - 3], "_"), splitted[l - 3], splitted[l - 2], splitted[l - 1], nil
}

