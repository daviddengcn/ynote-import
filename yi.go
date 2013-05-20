package main

import (
	"bufio"
	"code.google.com/p/mahonia"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/daviddengcn/go-ljson-conf"
	"github.com/daviddengcn/go-villa"
	ynote "github.com/daviddengcn/go-ynote"
	"html/template"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
)

const ac_FILENAME = "at.json"

func accFilePath() villa.Path {
	fn := villa.Path(ac_FILENAME)
	// Try user-home folder
	cu, err := user.Current()
	if err == nil {
		fn = villa.Path(cu.HomeDir).Join(ac_FILENAME)
	}

	return fn
}

func saveAccToken(ac *ynote.Credentials) {
	js, err := json.Marshal(ac)
	if err != nil {
		fmt.Println("Marshal accToken failed:", err)
		return
	}
	err = accFilePath().WriteFile(js, 0)
	if err != nil {
		fmt.Println("Write accToken failed:", err)
	}
}

func readAccToken() *ynote.Credentials {

	js, err := accFilePath().ReadFile()
	if err != nil {
		return nil
	}

	var cred ynote.Credentials
	err = json.Unmarshal(js, &cred)
	if err != nil {
		return nil
	}
	return &cred
}

func importDir(yc *ynote.YnoteClient, nbPath, dir villa.Path) error {
	fmt.Println("Importing files in folder", dir, " ... ")
	base := dir.Base()

	// Find or create the notebook.
	nbInfo, err := yc.CreateNotebook(base.S())
	if err != nil {
		nbInfo, err = yc.FindNotebook(base.S())
		if err != nil {
			fmt.Println("FindNotebook failed:", err)
			return err
		}
		fmt.Println("Folder", base, "found!")
	} else {
		fmt.Println("Folder", base, "created!")
	}

	files, err := dir.ReadDir()
	if err != nil {
		return err
	}

	for _, info := range files {
		if info.IsDir() {
			fmt.Println("  Ignoring folder:", info.Name())
			continue
		}
		nPath, err := importFile(yc, nbInfo.Path, dir.Join(info.Name()))
		if err != nil {
			return err
		}
		fmt.Println("imported:", nPath)
	}

	return nil
}

func text2html(text string) string {
	return strings.Replace(strings.Replace(strings.Replace(template.HTMLEscapeString(text), "\n", "<br>", -1), " ", "&nbsp;", -1), "\t", "&nbsp;&nbsp;&nbsp;&nbsp;", -1)
}

var gDecoder mahonia.Decoder
var gAuthor string
var gSource string
var gEncoding string

func init() {
	flag.StringVar(&gAuthor, "author", "GO-IMPORTER", "The author of imported notes.")
	flag.StringVar(&gSource, "source", "", "The source of imported notes.")
	flag.StringVar(&gEncoding, "enc", "utf-8", "The encoding of the input text.")
}

func importFile(yc *ynote.YnoteClient, nbPath string, fn villa.Path) (string, error) {
	fmt.Print("Importing ", fn, " ... ")
	content, err := fn.ReadFile()
	if err != nil {
		return "", err
	}

	//html := html(string(content))
	text := string(content)
	if gDecoder != nil {
		text = gDecoder.ConvertString(text)
	}
	html := text2html(text)

	return yc.CreateNote(nbPath, string(fn), gAuthor, gSource, html)
}

func initAfterParse() {
	if gEncoding != "utf-8" && gEncoding != "utf8" {
		gDecoder = mahonia.NewDecoder(gEncoding)
		if gDecoder == nil {
			fmt.Println("Unknown encoding", gEncoding+", supposing UTF-8")
		} else {
			fmt.Println("Supposing input encoding:", gEncoding)
		}
	}
}

func main() {
	flag.Parse()

	initAfterParse()

	conf, _ := ljconf.Load("yi.conf")
	yc := ynote.NewOnlineYnoteClient(ynote.Credentials{
		Token:  conf.String("key.token", ""),
		Secret: conf.String("key.secret", "")})

	yc.AccToken = readAccToken()
	if yc.AccToken == nil {
		fmt.Println("Access token (" + ac_FILENAME + ") not found, try authorize...")
		fmt.Println("Requesting temporary credentials ...")
		tmpCred, err := yc.RequestTemporaryCredentials()
		if err != nil {
			fmt.Println("RequestTemporaryCredentials failed: ", err)
			return
		}
		fmt.Println("Temporary credentials got:", tmpCred)

		authUrl := yc.AuthorizationURL(tmpCred)
		fmt.Println(authUrl)
		switch runtime.GOOS {
		case "darwin":
			exec.Command("open", authUrl).Start()
		case "windows":
			exec.Command("cmd", "/c", "start", authUrl).Start()
		case "linux":
			exec.Command("xdg-open", authUrl).Start()
		}

		fmt.Print("Please input the verifier: ")
		verifier, err := bufio.NewReader(os.Stdin).ReadString('\n')

		if err != nil {
			fmt.Println("Read verifier from console failed: ", err)
			return
		}

		verifier = strings.TrimSpace(verifier)
		fmt.Println("verifier:", verifier)

		accToken, err := yc.RequestToken(tmpCred, verifier)
		if err != nil {
			fmt.Println("RequestToken failed: ", err)
			return
		}

		fmt.Println(accToken)
		saveAccToken(accToken)
	}

	ui, err := yc.UserInfo()
	if err != nil {
		fmt.Println("Getting UserInfo failed:", err)
		return
	}
	fmt.Println("Hi,", ui.User)
	/*
		nbs, err := yc.ListNotebooks()
		if err != nil {
			fmt.Println("ListNotebooks failed:", err)
			return
		}
		fmt.Printf("%+v\n", nbs)
	//*/

	for i := 0; i < flag.NArg(); i++ {
		fn := villa.Path(flag.Arg(i)).AbsPath()

		if fn.IsDir() {
			err := importDir(yc, "", fn)
			if err != nil {
				fmt.Println("importDir failed:", err)
				continue
			}
		} else {
			pth, err := importFile(yc, "", fn)
			if err != nil {
				fmt.Println("importFile failed:", err)
				continue
			}
			fmt.Println("imported:", pth)
		}
	}
}
