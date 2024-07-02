package main

import (
    "flag"
    "fmt"
    "io/ioutil"
    "net"
    "os"
    "path/filepath"
    "strings"
)

var baseDir string

func handleConnection(conn net.Conn) {
    defer conn.Close()

    for {
        req := make([]byte, 1024)
        n, err := conn.Read(req)
        if err != nil || n == 0 {
            fmt.Println("Error reading from connection or connection closed")
            return
        }

        splitHeader := strings.Split(string(req[:n]), "\r\n")
        if len(splitHeader) < 1 {
            conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
            return
        }

        splitRequestLine := strings.Split(splitHeader[0], " ")
        if len(splitRequestLine) < 2 {
            conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
            return
        }

        path := splitRequestLine[1]

        if path == "/" {
            conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
        } else if strings.HasPrefix(path, "/echo/") {
            requestBody := strings.TrimPrefix(path, "/echo/")
            response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s",
                len(requestBody), requestBody)
            conn.Write([]byte(response))
        } else if path == "/user-agent" {
            var userAgent string
            for _, header := range splitHeader {
                if strings.HasPrefix(strings.ToLower(header), "user-agent:") {
                    userAgent = strings.TrimSpace(strings.SplitN(header, ":", 2)[1])
                    break
                }
            }

            if userAgent != "" {
                response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s",
                    len(userAgent), userAgent)
                conn.Write([]byte(response))
            } else {
                conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
            }
        } else if strings.HasPrefix(path, "/files/") {
            filename := strings.TrimPrefix(path, "/files/")
            filePath := filepath.Join(baseDir, filename)
            fileContent, err := ioutil.ReadFile(filePath)
            if err != nil {
                conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
                return
            }

            response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s",
                len(fileContent), fileContent)
            conn.Write([]byte(response))
        } else {
            conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
        }
    }
}

func main() {
    flag.StringVar(&baseDir, "directory", "", "The directory to serve files from")
    flag.Parse()

    if baseDir == "" {
        fmt.Println("Directory not specified")
        os.Exit(1)
    }

    fmt.Println("Logs from your program will appear here!")
    fmt.Printf("Serving files from directory: %s\n", baseDir)

    l, err := net.Listen("tcp", "0.0.0.0:4221")
    if err != nil {
        fmt.Println("Failed to bind to port 4221")
        os.Exit(1)
    }

    for {
        conn, err := l.Accept()
        if err != nil {
            fmt.Println("Error accepting connection: ", err.Error())
            os.Exit(1)
        }

        go handleConnection(conn)
    }
}
