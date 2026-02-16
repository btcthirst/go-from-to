package main

import (
"archive/zip"
"encoding/xml"
"fmt"
"io"
"strings"
)

func main() {
zipFile, _ := zip.OpenReader("нарахування26.ods")
defer zipFile.Close()

for _, file := range zipFile.File {
if file.Name == "content.xml" {
rc, _ := file.Open()
data, _ := io.ReadAll(rc)
rc.Close()

decoder := xml.NewDecoder(strings.NewReader(string(data)))
depth := 0

for {
token, err := decoder.Token()
if err != nil {
break
}

switch se := token.(type) {
case xml.StartElement:
if depth < 3 {
fmt.Printf("%s<%s>\n", strings.Repeat("  ", depth), se.Name.Local)
}
depth++

if se.Name.Local == "table" {
for _, attr := range se.Attr {
if attr.Name.Local == "name" {
fmt.Printf("  Found table: %s\n", attr.Value)
}
}
}

if depth > 10 {
depth = 10
}
case xml.EndElement:
depth--
if depth < 0 {
depth = 0
}
}
}
break
}
}
}
