package main

import (
    "fmt"
    "os"
    "github.com/4ensiX/img2df/util"
    "io"
)

func main() {

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
    dir := "temp"
    err2 := os.Mkdir(dir, 0755)
    if err2 != nil {
        fmt.Println(err2)
        os.Exit(1)
    }
    wf, err3 := os.Create(dir + "/tmp.tar")
    if err3 != nil {
        fmt.Println(err3)
        os.Exit(1)
    }
    if _, err4 := io.Copy(wf, reader); err != nil {
        fmt.Println(err4)
        os.Exit(1)
    }
    wf.Close()
    reader.Close()
    rf, err := os.Open(dir + "/tmp.tar")
    dfcmd, layers := util.ReadTar(rf)
    rf.Close()
    cpcmd, extLayers := util.CheckImageLayer(dfcmd, layers)
    rf2, err := os.Open(dir + "/tmp.tar")
    util.ExtractFiles(rf2,extLayers,cpcmd,dir)
    rf.Close()
    err5 := os.Remove(dir + "/tmp.tar")
    if err5 != nil {
        fmt.Println(err5)
        os.Exit(1)
    }
    util.CreateDockerfile(dfcmd,cpcmd,extLayers)
}
