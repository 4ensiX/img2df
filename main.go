package main

import (
    "fmt"
    "os"
    "github.com/4ensiX/img2df/img2df"
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
    reader, err := img2df.SaveImage(id)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    rf := img2df.SaveAndOpenImageTar(reader)
    reader.Close()
    dfcmd, layers := img2df.ReadTar(rf)
    rf.Close()
    cpcmd, extLayers := img2df.CheckImageLayer(dfcmd, layers)
    img2df.ExtractFiles(extLayers,cpcmd)
    img2df.CreateDockerfile(dfcmd,cpcmd,extLayers)
}
