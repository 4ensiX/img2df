package main

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

func saveImage(id string) (io.ReadCloser, error) {
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


func readTar(tarfile io.ReadCloser) ([]string){
    tarReader := tar.NewReader(tarfile)

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
            //fmt.Printf("%s\n",name)
            rep := regexp.MustCompile(`([A-Fa-f0-9]{64})\.json`)
            //rep := regexp.MustCompile(`manifest`)
            if rep.MatchString(name) {
                    layers = readHashJson(tarReader)
                    break
            }
    }
    return layers
}

type history struct {
//    Created string `json: "created"`
    Created_by string `json: "created_by"`
}

type conf struct {
    History []history `json: "history"`
}

func readHashJson(tarReader *tar.Reader) ([]string){

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


func formatRun(runc string) (string){
    var dline string
    
    dline = strings.Replace(runc, " \t", "\\ \n", -1)
    //dline = strings.Replace(runc, " \n", "\n", -1)
    dline = strings.Replace(dline, "\t&&", "&&", -1)

    return dline
}

func createDockerfile(commands []string){
    //FROM,RUN,CMD,LABEL,MAINTAINER,EXPOSE,ENV,ADD,COPY,ENTRYPOINT,VOLUME,USER,WORKDIR,ARG,ONBUILD,STOPSIGNAL,HEALTHCHECK,SHELL
    wf, err := os.Create("Dockerfile")
    if err != nil {
        fmt.Println(err)
        return
    }
    defer wf.Close()
    wf.WriteString("FROM scratch\n")
    wf.WriteString("\n")
    for _,tmp := range commands {
        if strings.HasPrefix(tmp, "/bin/sh -c") && !strings.Contains(tmp, "#(nop)") {//RUN
            runc := strings.Replace(tmp, "/bin/sh -c", "RUN", 1)
            //runc2 := strings.Replace(runc, " \t", "\\ \n", -1)
            runc2 := formatRun(runc)
            wf.WriteString(runc2 + "\n")
        }else if strings.Contains(tmp, "ADD") || strings.Contains(tmp, "COPY") { //ADD,COPY
            addc := strings.Replace(tmp, "/bin/sh -c #(nop) ", "", 1) // 1 space
            addc2 := strings.Replace(addc, "in", "", 1)
            wf.WriteString(addc2 + "\n")
        }else {
            docc := strings.Replace(tmp, "/bin/sh -c #(nop)  ", "", 1) // 2 space
            wf.WriteString(docc + "\n")
        }
        wf.WriteString("\n")
    }
}

func main(){
    id:= "python"
    reader, err := saveImage(id)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    defer reader.Close()

    commands:= readTar(reader)

    fmt.Println(commands)

    createDockerfile(commands)
}
