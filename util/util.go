package util

import (
    "fmt"
    "regexp"
    "os"
    "io"
    "archive/tar"
    "encoding/json"
    "io/ioutil"
    "strings"

    "github.com/docker/docker/client"
    "golang.org/x/net/context"

)

func SaveImage(id string) (io.ReadCloser, error) {
        var err error
        var cli *client.Client

        ctx := context.Background()

        cli, err = client.NewClientWithOpts(client.FromEnv)
        if err != nil {
                return nil, err
        }

        readCloser, err := cli.ImageSave(ctx,[]string{id})
        if err != nil {
                return nil, err
        }

        return readCloser, nil
}


type history struct {
    Created_by string `json: "created_by"`
}

type conf struct {
    History []history `json: "history"`
}

func ReadHashJson(tarReader *tar.Reader) ([]string){

    jsonfile, err := ioutil.ReadAll(tarReader)
    if err != nil {
            fmt.Println(err)
            os.Exit(1)
    }

    var l conf

    if err := json.Unmarshal([]byte(jsonfile), &l); err != nil {
            panic(err)
    }

    history := l.History

    var commands []string
    for _,com := range history {
            commands = append(commands,com.Created_by)
    }
    return commands

}

type manifest struct {
    Layers []string `json: "layers"`
}

func ReadManifest(tarReader *tar.Reader) ([]string){

    jsonfile, err := ioutil.ReadAll(tarReader)
    if err != nil {
            fmt.Println(err)
            os.Exit(1)
    }

    var m []manifest
    if err := json.Unmarshal([]byte(jsonfile), &m); err != nil {
            panic(err)
    }
    m1 := m[0]
    var layers []string
    for _,i := range m1.Layers {
        layers = append(layers,strings.TrimRight(i, "/layer.tar"))
    }
    return layers
}

func ReadTar(tarfile io.ReadCloser) ([]string,[]string){
    tarReader := tar.NewReader(tarfile)

    var dfcmd []string
    var layers []string

    for {
            tarHeader, err := tarReader.Next()
            if err == io.EOF {
                    break
            }

            if err != nil {
                    fmt.Println(err)
                    os.Exit(1)
            }

            name := tarHeader.Name
            rep := regexp.MustCompile(`([A-Fa-f0-9]{64})\.json`)
            if rep.MatchString(name) {
                    dfcmd = ReadHashJson(tarReader)
            }else if strings.HasPrefix(name,"manifest.json"){
                    layers = ReadManifest(tarReader)
            }
    }
    return dfcmd,layers
}

func FormatRun(runc string) (string){
    var dline string

    dline = strings.Replace(runc, " \t", "\\ \n", -1)
    dline = strings.Replace(dline, "\t&&", "&&", -1)

    return dline
}

func CheckImageLayer(dfcmd []string, layers []string) ([]string,[]string){//dfcmd->dockerfile lines,layers->image layers
    var imgcmd = []string{"/bin/sh -c #(nop) ADD","/bin/sh -c #(nop) COPY","/bin/sh -c #(nop) WORKDIR"}
    var imgLayerCmd []string
    for _,i := range dfcmd {
        if strings.Contains(i,imgcmd[0]) || strings.Contains(i,imgcmd[1]) || strings.Contains(i,imgcmd[2]){
            imgLayerCmd = append(imgLayerCmd,i)
        }else if strings.HasPrefix(i, "/bin/sh -c") && !strings.Contains(i, "#(nop)") {//RUN
            imgLayerCmd = append(imgLayerCmd,i)
        }
    }
    if len(imgLayerCmd) != len(layers){
        fmt.Println("create image-layer commands and image layers is different?")
        os.Exit(1)
    }
    var cpcmd []string
    var layerFiles []string
    for i,j := range imgLayerCmd {
        if strings.Contains(j,imgcmd[0]) || strings.Contains(j,imgcmd[1]) {
            cpcmd = append(cpcmd,j)
            layerFiles = append(layerFiles,layers[i])
        }
    }
    return cpcmd,layerFiles
}

func CheckLayer(name string, layers []string) (int){
    for i,j := range layers {
        tmp := j + "/layer.tar"
        if strings.HasPrefix(name,tmp){return i}
    }
    return -1
}

func CopyLayer(dir string, tarfile *tar.Reader) {
    tarReader := tar.NewReader(tarfile)
    for {
            tarHeader, err := tarReader.Next()
            if err == io.EOF {
                    break
            }

            if err != nil {
                    fmt.Println(err)
                    os.Exit(1)
            }

            name := tarHeader.Name
            if tarHeader.Typeflag == tar.TypeDir {
                err = os.Mkdir(dir + "/" + name, 0755)
                if err != nil {
                   fmt.Println(err)
                    os.Exit(1)
                }
            }else if tarHeader.Typeflag == tar.TypeSymlink || tarHeader.Typeflag == tar.TypeReg {
                wf, err := os.Create(dir + "/" + name)
                if err != nil {
                   fmt.Println(err)
                   os.Exit(1)
                }
                if _, err := io.Copy(wf, tarReader); err != nil {
                   fmt.Println(err)
                   os.Exit(1)
                }
                wf.Close()
            }
    }
}

func CopyFiles(tarfile *tar.Reader, layer string, cpcmd string , dir string) {
    if strings.HasPrefix(cpcmd,"/bin/sh -c #(nop) ADD") { 
        wf, err := os.Create(dir + "/" + layer + ".tar")
        if err != nil {
            fmt.Println(err)
            os.Exit(1)
        }
        if _, err := io.Copy(wf, tarfile); err != nil {
            fmt.Println(err)
            os.Exit(1)
        }
        wf.Close()
    }else if strings.HasPrefix(cpcmd,"/bin/sh -c #(nop) COPY") {
        if err := os.Mkdir(dir + "/" + layer, 0755); err != nil {
            fmt.Println(err)
            os.Exit(1)
        }
        CopyLayer(dir + "/" + layer, tarfile)
    }
}

func ExtractFiles(tarfile io.Reader, layers []string, cpcmds []string, dir string) {

    var il int = 0
    tarReader := tar.NewReader(tarfile)

    for {
            tarHeader, err := tarReader.Next()
            if err == io.EOF {
                    break
            }
            if err != nil {
                    fmt.Println(err)
                    os.Exit(1)
            }

            name := tarHeader.Name
            il = CheckLayer(name,layers)
            if il >= 0 {
                CopyFiles(tarReader,layers[il],cpcmds[il],dir)
            }
    }
}


func FormatCPAD(addc string, cpcmds []string, extLayers []string) (string){
    var il int = 0
    for i,j := range cpcmds {
        if strings.HasPrefix(addc,j){il = i}
    }
    addc2 := strings.Replace(addc, "/bin/sh -c #(nop) ", "", 1) // 1 space
    addc3 := strings.Replace(addc2, "in ", "", 1)
    addcs := strings.Split(addc3, " ")
    var tmp string
    tmp = "temp" + "/" + extLayers[il]
    if strings.HasPrefix(addcs[0],"ADD") {
        tmp = tmp + ".tar"
    }else {//COPY
        tmp = tmp + addcs[2]
    }
    return addcs[0] + " " + tmp + " " + addcs[2]
}

func CreateDockerfile(dfcmds []string, cpcmds []string, extLayers []string){
    //FROM,RUN,CMD,LABEL,MAINTAINER,EXPOSE,ENV,ADD,COPY,ENTRYPOINT,VOLUME,USER,WORKDIR,ARG,ONBUILD,STOPSIGNAL,HEALTHCHECK,SHELL
    wf, err := os.Create("Dockerfile")
    if err != nil {
        fmt.Println(err)
        return
    }
    defer wf.Close()
    wf.WriteString("FROM scratch\n")
    wf.WriteString("\n")
    for _,tmp := range dfcmds {
        if strings.HasPrefix(tmp, "/bin/sh -c") && !strings.Contains(tmp, "#(nop)") {//RUN
            runc := strings.Replace(tmp, "/bin/sh -c", "RUN", 1)
            runc2 := FormatRun(runc)
            wf.WriteString(runc2 + "\n")
        }else if strings.Contains(tmp, "ADD") || strings.Contains(tmp, "COPY") { //ADD,COPY
            addc := FormatCPAD(tmp,cpcmds,extLayers)
            wf.WriteString(addc + "\n")
        }else if strings.Contains(tmp, "LABEL") { //LABEL
            labc := strings.Replace(tmp, "/bin/sh -c #(nop)  ", "", 1)
            labc2 := strings.Split(labc, "=")
            tmp2 := labc2[0] + "=" + "\"" + labc2[1] + "\""
            wf.WriteString(tmp2 + "\n")
        }else {
            docc := strings.Replace(tmp, "/bin/sh -c #(nop)  ", "", 1) // 2 space
            wf.WriteString(docc + "\n")
        }// WORKDIR
        wf.WriteString("\n")
    }
}
