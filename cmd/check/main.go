package main

import (
"archive/zip"
"fmt"
"io"
)

func main() {
fmt.Println("=== АНАЛІЗ СТРУКТУРИ ODS ===")
zipFile, err := zip.OpenReader("нарахування26.ods")
if err != nil {
fmt.Printf("Помилка: %v\n", err)
return
}
defer zipFile.Close()

fmt.Println("Файли в архіві:")
for _, file := range zipFile.File {
fmt.Printf("  - %s\n", file.Name)
}

fmt.Println("\n=== ПЕРШІ 3000 СИМВОЛІВ content.xml ===")
for _, file := range zipFile.File {
if file.Name == "content.xml" {
rc, _ := file.Open()
data, _ := io.ReadAll(rc)
rc.Close()

content := string(data)
if len(content) > 3000 {
fmt.Println(content[:3000])
} else {
fmt.Println(content)
}
break
}
}
}
