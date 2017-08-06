package files

import "os"

func IsDir(path string) int {
    info, err := os.Stat(path)
    if err!=nil{
        return -1
    }
    if info.IsDir(){
        return 1
    }else {
        return 0
    }
}

func Exist(path string) bool {
    _, err := os.Stat(path)
    return err == nil || os.IsExist(err)
}