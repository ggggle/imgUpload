package main

import (
    "os"
    "gopkg.in/alecthomas/kingpin.v2"
    "fmt"
    "net/http"
    "bytes"
    "mime/multipart"
    "io"
    "io/ioutil"
    "github.com/buger/jsonparser"
    "strings"
    "./files"
)

var (
    app     = kingpin.New("imgUpload", "imgUpload")
    imgPath = app.Arg("imgPath", "imgPath").Required().Strings()
)

func main() {
    kingpin.MustParse(app.Parse(os.Args[1:]))
    fmt.Println(*imgPath)
    for _, arg := range *imgPath {
        switch files.IsDir(arg) {
        case -1:
            fmt.Println("文件不存在")
        case 0: //单图片上传
            response, err := upload(arg)
            if err == nil {
                dealResponse(response, arg)
            } else {
                fmt.Println("%v", err)
            }
        case 1: //文件夹遍历图片上传

        }
    }
    var a string
    fmt.Scanf("%s", &a)
}

func upload(fileName string) (*http.Response, error) {
    body_buf := bytes.NewBufferString("")
    body_writer := multipart.NewWriter(body_buf)
    fileWriter, err := body_writer.CreateFormFile("smfile", fileName)
    if err != nil {
        fmt.Println("error while CreateFormFile")
        return nil, err
    }
    fh, err := os.Open(fileName)
    if err != nil {
        fmt.Println("error opening file")
        return nil, err
    }
    defer fh.Close()
    _, err = io.Copy(fileWriter, fh)
    if err != nil {
        return nil, err
    }
    contentType := body_writer.FormDataContentType()
    body_writer.Close()
    resp, err := http.Post("https://sm.ms/api/upload", contentType, body_buf)
    return resp, err
}

func dealResponse(resp *http.Response, fileName string) {
    defer resp.Body.Close()
    resp_body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return
    }
    if resp.StatusCode != 200 {
        fmt.Println("http状态码[%d]", resp.StatusCode)
        return
    }
    jsonData := string(resp_body)
    url, _ := jsonparser.GetString([]byte(jsonData), "data", "url")
    fmt.Println(url)
    delete, _ := jsonparser.GetString([]byte(jsonData), "data", "delete")
    path, _ := jsonparser.GetString([]byte(jsonData), "data", "path")
    if len(url) > 0 {
        linkFile, err := os.OpenFile("link.txt", os.O_CREATE|os.O_APPEND, 0644)
        defer linkFile.Close()
        if err != nil {
            fmt.Println("open link.txt error %v", err)
        } else {
            linkFile.WriteString(strings.Join([]string{url, delete}, ",") + "\n")
        }
    }
    logFile, err := os.OpenFile("log.txt", os.O_CREATE|os.O_APPEND, 0644)
    if err != nil {
        fmt.Println("open log.txt error %v", err)
    } else {
        logFile.WriteString(strings.Replace(jsonData, "\\", "", -1) + "\n")
    }
    path = "." + path
    pathSplit := strings.Split(path, "/")
    dict := strings.Join(pathSplit[0:len(pathSplit)-1], string(os.PathSeparator))
    os.MkdirAll(dict, 0644)
    CopyFile(fileName, path)
}

// CopyFile copies a file from src to dst. If src and dst files exist, and are
// the same, then return success. Otherise, attempt to create a hard link
// between the two files. If that fail, copy the file contents from src to dst.
func CopyFile(src, dst string) (err error) {
    sfi, err := os.Stat(src)
    if err != nil {
        return
    }
    if !sfi.Mode().IsRegular() {
        // cannot copy non-regular files (e.g., directories,
        // symlinks, devices, etc.)
        return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
    }
    dfi, err := os.Stat(dst)
    if err != nil {
        if !os.IsNotExist(err) {
            return
        }
    } else {
        if !(dfi.Mode().IsRegular()) {
            return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
        }
        if os.SameFile(sfi, dfi) {
            return
        }
    }
    if err = os.Link(src, dst); err == nil {
        return
    }
    err = copyFileContents(src, dst)
    return
}

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func copyFileContents(src, dst string) (err error) {
    in, err := os.Open(src)
    if err != nil {
        return
    }
    defer in.Close()
    out, err := os.Create(dst)
    if err != nil {
        return
    }
    defer func() {
        cerr := out.Close()
        if err == nil {
            err = cerr
        }
    }()
    if _, err = io.Copy(out, in); err != nil {
        return
    }
    err = out.Sync()
    return
}
