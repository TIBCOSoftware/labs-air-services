package util

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/msoap/byline"
)

type DepManager interface {
	Init() error
	AddDependency(flogoImport Import) error
	GetPath(flogoImport Import) (string, error)
	AddReplacedContribForBuild() error
	InstallReplacedPkg(string, string) error
	GetAllImports() (map[string]Import, error)
}

func NewDepManager(sourceDir string) DepManager {
	fmt.Printf("(mod.NewDepManager) Entering : sourceDir=%s\n", sourceDir)
	defer fmt.Println("(mod.NewDepManager) Exiting ....")
	return &ModDepManager{srcDir: sourceDir, localMods: make(map[string]string)}
}

type ModDepManager struct {
	srcDir    string
	localMods map[string]string
}

func (m *ModDepManager) Init() error {
	fmt.Printf("(mod.ModDepManager.Init) Entering Entering ....")
	defer fmt.Println("(mod.ModDepManager.Init) Exiting ....")

	err := ExecCmd(exec.Command("go", "mod", "init", "main"), m.srcDir)
	if err == nil {
		return err
	}

	return nil
}

func (m *ModDepManager) AddDependency(flogoImport Import) error {
	fmt.Printf("(mod.ModDepManager.AddDependency) Entering : flogoImport=%v\n", flogoImport)
	defer fmt.Println("(mod.ModDepManager.AddDependency) Exiting ....")

	// todo: optimize the following

	// use "go mod edit" (instead of "go get") as first method
	err := ExecCmd(exec.Command("go", "mod", "edit", "-require", flogoImport.GoModImportPath()), m.srcDir)
	if err != nil {
		return err
	}

	err = ExecCmd(exec.Command("go", "mod", "verify"), m.srcDir)
	if err == nil {
		err = ExecCmd(exec.Command("go", "mod", "download", flogoImport.ModulePath()), m.srcDir)
	}

	if err != nil {
		// if the resolution fails and the Flogo import is "classic"
		// (meaning it does not separate module path from Go import path):
		// 1. remove the import manually ("go mod edit -droprequire") would fail
		// 2. try with "go get" instead
		if flogoImport.IsClassic() {
			m.RemoveImport(flogoImport)
			err = ExecCmd(exec.Command("go", "get", flogoImport.GoGetImportPath()), m.srcDir)
		}
	}

	if err != nil {
		return err
	}

	return nil
}

// GetPath gets the path of where the
func (m *ModDepManager) GetPath(flogoImport Import) (string, error) {
	fmt.Printf("(mod.ModDepManager.GetPath) Entering : flogoImport=%s\n", flogoImport)
	defer fmt.Println("(mod.ModDepManager.GetPath) Exiting ....")

	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	fmt.Printf("(mod.ModDepManager.GetPath) pass 1 : currentDir=%s\n", currentDir)

	pkg := flogoImport.ModulePath()

	path, ok := m.localMods[pkg]
	if ok && path != "" {

		return path, nil
	}
	fmt.Printf("(mod.ModDepManager.GetPath) pass 2 : path=%s\n", path)
	defer os.Chdir(currentDir)

	os.Chdir(m.srcDir)

	file, err := os.Open(filepath.Join(m.srcDir, "go.mod"))
	defer file.Close()

	var pathForPartial string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {

		line := scanner.Text()
		reqComponents := strings.Fields(line)
		fmt.Printf("(mod.ModDepManager.GetPath) pass 3 : reqComponents=%v\n", reqComponents)
		//It is the line in the go.mod which is not useful, so ignore.
		if len(reqComponents) < 2 || (reqComponents[0] == "require" && reqComponents[1] == "(") {
			continue
		}

		//typically package is 1st component and  version is the 2nd component
		reqPkg := reqComponents[0]
		version := reqComponents[1]
		fmt.Printf("(mod.ModDepManager.GetPath) pass 4 : pkg=%s reqPkg=%s version=%s\n", pkg, reqPkg, version)
		if reqComponents[0] == "require" {
			//starts with require, so package is 2nd component and  version is the 3rd component
			reqPkg = reqComponents[1]
			version = reqComponents[2]
		}

		if strings.HasPrefix(pkg, reqPkg) {
			fmt.Printf("(mod.ModDepManager.GetPath) pass 5 ....\n")

			hasFull := strings.Contains(line, pkg)
			tempPath := strings.Split(reqPkg, "/")

			tempPath = toLower(tempPath)
			lastIdx := len(tempPath) - 1

			tempPath[lastIdx] = tempPath[lastIdx] + "@" + version

			pkgPath := filepath.Join(tempPath...)

			if !hasFull {
				remaining := pkg[len(reqPkg):]
				tempPath = strings.Split(remaining, "/")
				remainingPath := filepath.Join(tempPath...)

				pathForPartial = filepath.Join(os.Getenv("GOPATH"), "pkg", "mod", pkgPath, remainingPath)
			} else {
				return filepath.Join(os.Getenv("GOPATH"), "pkg", "mod", pkgPath, flogoImport.RelativeImportPath()), nil
			}
		}
	}
	return pathForPartial, nil
}

func (m *ModDepManager) RemoveImport(flogoImport Import) error {
	fmt.Printf("(mod.ModDepManager.RemoveImport) Entering : flogoImport=%v\n", flogoImport)
	defer fmt.Println("(mod.ModDepManager.RemoveImport) Exiting ....")

	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}

	modulePath := flogoImport.ModulePath()

	defer os.Chdir(currentDir)

	os.Chdir(m.srcDir)

	file, err := os.Open(filepath.Join(m.srcDir, "go.mod"))
	if err != nil {
		return err
	}
	defer file.Close()

	modulePath = strings.Replace(modulePath, "/", "\\/", -1)
	modulePath = strings.Replace(modulePath, ".", "\\.", -1)
	importRegex := regexp.MustCompile(`\s*` + modulePath + `\s+` + flogoImport.Version() + `.*`)

	lr := byline.NewReader(file)

	lr.MapString(func(line string) string {
		if importRegex.MatchString(line) {
			return ""
		} else {
			return line
		}
	})

	updatedGoMod, err := lr.ReadAll()
	if err != nil {
		return err
	}

	file, err = os.Create(filepath.Join(m.srcDir, "go.mod"))
	if err != nil {
		return err
	}
	defer file.Close()

	file.Write(updatedGoMod)

	return nil
}
func (m *ModDepManager) GetAllImports() (map[string]Import, error) {
	fmt.Printf("(mod.ModDepManager.GetAllImports) Entering ....")
	defer fmt.Println("(mod.ModDepManager.GetAllImports) Exiting ....")
	file, err := ioutil.ReadFile(filepath.Join(m.srcDir, "go.mod"))
	if err != nil {
		return nil, err
	}

	content := string(file)

	imports := strings.Split(content[strings.Index(content, "(")+1:strings.Index(content, ")")], "\n")
	result := make(map[string]Import)

	for _, pkg := range imports {
		if pkg != " " && pkg != "" {

			mods := strings.Split(strings.TrimSpace(pkg), " ")

			modImport, err := ParseImport(strings.Join(mods[:2], "@"))
			if err != nil {
				return nil, err
			}

			result[modImport.GoImportPath()] = modImport
		}
	}

	return result, nil
}

//This function converts capotal letters in package name
// to !(smallercase). Eg C => !c . As this is the way
// go.mod saves every repository in the $GOPATH/pkg/mod.
func toLower(s []string) []string {
	fmt.Printf("(mod.toLower) Entering : s=%v\n", s)
	defer fmt.Println("(mod.toLower) Exiting ....")
	result := make([]string, len(s))
	for i := 0; i < len(s); i++ {
		var b bytes.Buffer
		for _, c := range s[i] {
			if c >= 65 && c <= 90 {
				b.WriteRune(33)
				b.WriteRune(c + 32)
			} else {
				b.WriteRune(c)
			}
		}
		result[i] = b.String()
	}
	return result
}

var verbose = false

func SetVerbose(enable bool) {
	verbose = enable
}

func Verbose() bool {
	return verbose
}

func ExecCmd(cmd *exec.Cmd, workingDir string) error {
	fmt.Printf("(mod.ExecCmd) Entering : cmd=%v, workingDir=%s\n", cmd, workingDir)
	defer fmt.Println("(mod.ExecCmd) Exiting ....")

	if workingDir != "" {
		cmd.Dir = workingDir
	}

	var out bytes.Buffer

	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdout = nil
		cmd.Stderr = &out
	}

	err := cmd.Run()

	if err != nil {
		return fmt.Errorf(string(out.Bytes()))
	}

	return nil
}

func (m *ModDepManager) AddReplacedContribForBuild() error {
	fmt.Printf("(mod.ModDepManager.AddReplacedContribForBuild) Entering : m.srcDir = %v\n", m.srcDir)
	defer fmt.Println("(mod.ModDepManager.AddReplacedContribForBuild) Exiting ....")

	err := ExecCmd(exec.Command("go", "mod", "download"), m.srcDir)
	if err != nil {
		return err
	}

	text, err := ioutil.ReadFile(filepath.Join(m.srcDir, "go.mod"))
	if err != nil {
		return err
	}

	data := string(text)

	index := strings.Index(data, "replace")
	if index != -1 {
		localModules := strings.Split(data[index-1:], "\n")

		fmt.Printf("(mod.ModDepManager.AddReplacedContribForBuild) m.localMods = %s\n", m.localMods)
		for _, val := range localModules {
			fmt.Printf("(mod.ModDepManager.AddReplacedContribForBuild) val = %s\n", val)
			if val != "" {
				mods := strings.Split(val, " ")
				//If the length of mods is more than 4 it contains the versions of package
				//so it is stating to use different version of pkg rather than
				// the local pkg.
				fmt.Printf("(mod.ModDepManager.AddReplacedContribForBuild) mods = %s\n", mods)
				if len(mods) >= 4 {
					if len(mods) < 5 {
						m.localMods[mods[1]] = mods[3]
					} else {

						m.localMods[mods[1]] = filepath.Join(os.Getenv("GOPATH"), "pkg", "mod", mods[3]+"@"+mods[4])
					}
				} else {
					fmt.Printf("(mod.ModDepManager.AddReplacedContribForBuild) len(mods) < 4 \n")
				}
			}
		}
		return nil
	}
	return nil
}

func (m *ModDepManager) InstallReplacedPkg(pkg1 string, pkg2 string) error {
	fmt.Printf("(mod.ModDepManager.InstallReplacedPkg) Entering : pkg1=%s, pkg2=%s\n", pkg1, pkg2)
	defer fmt.Println("(mod.ModDepManager.InstallReplacedPkg) Exiting ....")

	m.localMods[pkg1] = pkg2

	f, err := os.OpenFile(filepath.Join(m.srcDir, "go.mod"), os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err = f.WriteString(fmt.Sprintf("replace %v => %v", pkg1, pkg2)); err != nil {
		return err
	}

	err = ExecCmd(exec.Command("go", "mod", "download"), m.srcDir)
	if err != nil {
		return err
	}
	return nil
}

type Resp struct {
	Name string `json:"name"`
}

func getLatestVersion(path string) string {
	fmt.Printf("(mod.getLatestVersion) Entering : path=%s\n", path)
	defer fmt.Println("(mod.getLatestVersion) Exiting ....")

	//To get the latest version number use the  GitHub API.
	resp, err := http.Get("https://api.github.com/repos/TIBCOSoftware/flogo-contrib/releases/latest")
	if err != nil {
		log.Fatalln(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var result Resp

	err = json.Unmarshal(body, &result)
	if err != nil {
		return ""
	}

	return result.Name

}
