package main

import (
    "fmt"
    "os"
    "github.com/4ensiX/img2df/util"
)

func main(){

    if len(os.Args) < 2 {
        fmt.Println("usage: img2df [image name] or [image:tag]")
        fmt.Println("example: img2df debian")
        return
    }else if len(os.Args) > 2 {
        fmt.Println("invalid args")
        return
    }
    var id string = os.Args[1]
    reader, err := util.SaveImage(id)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    defer reader.Close()

    commands:= util.ReadTar(reader)
    util.CreateDockerfile(commands)
}
