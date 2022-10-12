package main

import (
	"errors"
	"fmt"
	"go/build"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/urfave/cli/v2"
)

var (
	withBM        bool
	withGin       bool
	withGRPC      bool
	withTcp       bool
	withTcpLoop   bool
	withWS        bool
	withComet     bool
	withGomsg     bool
	withGomsgLoop bool
	withSwagger   bool
	withEcode     bool
)

func protocAction(ctx *cli.Context) (err error) {
	if err = checkProtoc(); err != nil {
		return err
	}
	files := ctx.Args().Slice()
	if len(files) == 0 {
		files, _ = filepath.Glob("*.proto")
	}
	//if !withGRPC && !withBM && !withSwagger && !withEcode {
	//	withBM = true
	//	withGRPC = true
	//	withSwagger = true
	//	withEcode = true
	//}
	if !withGRPC && !withGin && !withSwagger && !withEcode {
		withGin = true
		withGRPC = true
		withSwagger = true
		withEcode = true
	}
	if withTcp || withTcpLoop {
		withGin = true
		withGRPC = true
		withSwagger = true
		withEcode = true
	}
	if withWS {
		withGin = true
		withGRPC = true
		withSwagger = true
		withEcode = true
	}
	if withComet {
		withGin = true
		withSwagger = true
		withEcode = true
	}
	if withGomsg || withGomsgLoop {
		withGin = true
		withSwagger = true
		withEcode = true
	}
	if withComet {
		if err = genComet(files); err != nil {
			return
		}
	}
	if withGomsg {
		if err = genGoMsg(files); err != nil {
			return
		}
	}
	if withGomsgLoop {
		if err = genGoMsgLoop(files); err != nil {
			return
		}
	}
	if withTcp {
		if err = genTcp(files); err != nil {
			return
		}
	}
	if withTcpLoop {
		if err = genTcpLoop(files); err != nil {
			return
		}
	}
	if withWS {
		if err = genWS(files); err != nil {
			return
		}
	}

	if withBM {
		if err = installBMGen(); err != nil {
			return
		}
		if err = genBM(files); err != nil {
			return
		}
	}
	if withGin {
		if err = installGinGen(); err != nil {
			return
		}
		if err = genGin(files); err != nil {
			return
		}
	}
	if withGRPC {
		if err = installGRPCGen(); err != nil {
			return err
		}
		if err = genGRPC(files); err != nil {
			return
		}
	}
	if withSwagger {
		if err = installSwaggerGen(); err != nil {
			return
		}
		if err = genSwagger(files); err != nil {
			return
		}
	}
	if withEcode {
		if err = installEcodeGen(); err != nil {
			return
		}
		if err = genEcode(files); err != nil {
			return
		}
	}
	log.Printf("generate %s success.\n", strings.Join(files, " "))
	return nil
}

func checkProtoc() error {
	if _, err := exec.LookPath("protoc"); err != nil {
		switch runtime.GOOS {
		case "darwin":
			fmt.Println("brew install protobuf")
			cmd := exec.Command("brew", "install", "protobuf")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err = cmd.Run(); err != nil {
				return err
			}
		case "linux":
			fmt.Println("snap install --classic protobuf")
			cmd := exec.Command("snap", "install", "--classic", "protobuf")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err = cmd.Run(); err != nil {
				return err
			}
		default:
			return errors.New("您还没安装protobuf，请进行手动安装：https://github.com/protocolbuffers/protobuf/releases")
		}
	}
	return nil
}

func generate(protoc string, files []string) error {
	pwd, _ := os.Getwd()
	gosrc := path.Join(gopath(), "src")
	//ext, err := latestKratos()
	gogoproto, err := latestgogoproto()
	gogogapis, err := latestgogoapis()
	if err != nil {
		return err
	}
	line := fmt.Sprintf(protoc, gosrc, gogoproto, gogogapis, pwd)
	log.Println(line, strings.Join(files, " "))
	args := strings.Split(line, " ")
	args = append(args, files...)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = pwd
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func goget(url string) error {
	args := strings.Split(url, " ")
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Println(url)
	return cmd.Run()
}

func latestKratos() (string, error) {
	gopath := gopath()
	ext := path.Join(gopath, "src/github.com/peterlearn/kratos/v1/third_party")
	if _, err := os.Stat(ext); !os.IsNotExist(err) {
		return ext, nil
	}
	ext = path.Join(gopath, "src/kratos/third_party")
	if _, err := os.Stat(ext); !os.IsNotExist(err) {
		return ext, nil
	}
	baseMod := path.Join(gopath, "pkg/mod/git.huoys.com/middle-end")
	files, err := ioutil.ReadDir(baseMod)
	if err != nil {
		return "", err
	}
	for i := len(files) - 1; i >= 0; i-- {
		if strings.HasPrefix(files[i].Name(), "kratos@") {
			return path.Join(baseMod, files[i].Name(), "third_party"), nil
		}
	}
	return "", errors.New("not found kratos package")
}

func latestgogoproto() (string, error) {
	gopath := gopath()
	ext := path.Join(gopath, "src/github.com/gogo/protobuf")
	if _, err := os.Stat(ext); !os.IsNotExist(err) {
		return ext, nil
	}
	ext = path.Join(gopath, "src/gogo/protobuf")
	if _, err := os.Stat(ext); !os.IsNotExist(err) {
		return ext, nil
	}
	baseMod := path.Join(gopath, "pkg/mod/github.com/gogo")
	files, err := ioutil.ReadDir(baseMod)
	if err != nil {
		return "", err
	}
	for i := len(files) - 1; i >= 0; i-- {
		if strings.HasPrefix(files[i].Name(), "protobuf@") {
			return path.Join(baseMod, files[i].Name()), nil
		}
	}
	return "", errors.New("not found kratos package")
}

func latestgogoapis() (string, error) {
	gopath := gopath()
	ext := path.Join(gopath, "src/github.com/gogo/googleapis")
	if _, err := os.Stat(ext); !os.IsNotExist(err) {
		return ext, nil
	}
	ext = path.Join(gopath, "src/gogo/googleapis")
	if _, err := os.Stat(ext); !os.IsNotExist(err) {
		return ext, nil
	}
	baseMod := path.Join(gopath, "pkg/mod/github.com/gogo")
	files, err := ioutil.ReadDir(baseMod)
	if err != nil {
		return "", err
	}
	for i := len(files) - 1; i >= 0; i-- {
		if strings.HasPrefix(files[i].Name(), "googleapis@") {
			return path.Join(baseMod, files[i].Name()), nil
		}
	}
	return "", errors.New("not found kratos package")
}

func gopath() (gp string) {
	gopaths := strings.Split(os.Getenv("GOPATH"), string(filepath.ListSeparator))

	if len(gopaths) == 1 && gopaths[0] != "" {
		return gopaths[0]
	}
	pwd, err := os.Getwd()
	if err != nil {
		return
	}
	abspwd, err := filepath.Abs(pwd)
	if err != nil {
		return
	}
	for _, gopath := range gopaths {
		if gopath == "" {
			continue
		}
		absgp, err := filepath.Abs(gopath)
		if err != nil {
			return
		}
		if strings.HasPrefix(abspwd, absgp) {
			return absgp
		}
	}
	return build.Default.GOPATH
}
