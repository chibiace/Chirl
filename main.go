package main

import (
	"bufio"
	"embed"
	"flag"
	"fmt"
	"html/template"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

//go:embed all:templates
var FStemplates embed.FS

type Project struct {
	Name      string
	License   string
	Version   string
	Creator   string
	Directory string
	Year      string
}

// Lets Go!
func main() {
	// patriotism
	projectNameFlag := flag.String("name", "", "Project name: -name \"My Project\"")
	licenseNameFlag := flag.String("license", "", "Software license: -license MIT")
	versionFlag := flag.String("version", "", "Software version: -version \"1.0\"")
	creatorNameFlag := flag.String("creator", "", "Creator(s) aka copyright holders: -creator \"John Smith\"")
	viewLicenseFlag := flag.String("view", "", "Views a license: -view MIT")
	listLicensesFlag := flag.Bool("list", false, "Lists all the included licenses")
	flag.Parse()

	// view license if exists, exit.
	if len(*viewLicenseFlag) != 0 {
		if checkLicense(*viewLicenseFlag) {
			view(*viewLicenseFlag)
		} else {
			fmt.Println("License does not exist, use -list to see licenses")
			os.Exit(1)
		}
		os.Exit(1)
	}

	// list licenses, exit.
	if *listLicensesFlag {
		list()
		os.Exit(1)
	}

	// check directory otherwise default
	directoryArg := flag.Arg(0)
	directory := "./"
	if len(directoryArg) != 0 {
		if _, err := os.Stat(directoryArg); os.IsNotExist(err) {
			// specified directory doesn't exist
			panic(err)
		} else {
			if _, err := os.Stat(directory); os.IsNotExist(err) {
				// default directory doesn't exist
				panic(err)
			} else {
				directory = directoryArg
			}
		}
	}

	//checks the license if ones been specified by the flag argument
	if len(*licenseNameFlag) != 0 && !checkLicense(*licenseNameFlag) {
		fmt.Println("Unknown license try -list or -view licenseName ")
		os.Exit(1)
	}

	// partying like its 2006
	year := time.Now().Format("2006")

	//make the struct
	newProject := Project{*projectNameFlag, *licenseNameFlag, *versionFlag, *creatorNameFlag, directory, year}

	// time to move on
	form(&newProject)
}

// gets data from the user
func form(p *Project) {
	if len(p.Name) == 0 {
		p.Name = input("Project Name (default: My Project):", true, "My Project")
	}
	if len(p.License) == 0 {
		var a string
		for !checkLicense(a) {
			a = input("License (default: MIT) [list = list licenses]:", true, "MIT")
			if a == "list" {
				list()
			}
		}
		p.License = a
	}
	if len(p.Version) == 0 {
		p.Version = input("Version (default: 1.0.0):", true, "1.0.0")
	}
	if len(p.Creator) == 0 {
		osUser, _ := user.Current()
		username := osUser.Username
		p.Creator = input(fmt.Sprintf("Creator (default: %v):", username), true, username)
	}

	fmt.Println("-----------------------")
	fmt.Println("Directory:", p.Directory)
	fmt.Println("Project Name:", p.Name)
	fmt.Println("License:", p.License)
	fmt.Println("Version:", p.Version)
	fmt.Println("Creator:", p.Creator)
	fmt.Println("Year:", p.Year)
	fmt.Println("-----------------------")
	complete := inputYesNo("Happy with this?")
	if complete {
		//do stuff
		createFiles(p)
	} else {
		fmt.Println("Exiting.. Goodbye!")
	}
}

// checks to see if a license exists
func checkLicense(a string) bool {
	files, err := FStemplates.ReadDir("templates/licenses")
	if err != nil {
		panic(err)
	}
	var licenses []string
	for _, f := range files {
		licenses = append(licenses, f.Name())
	}
	for _, license := range licenses {
		if license == a {
			return true
		}
	}
	return false
}

// user input
func input(q string, allowEmpty bool, defaultValue ...string) string {
	var a string
	for len(a) == 0 {
		in := bufio.NewReader(os.Stdin)
		fmt.Print(q + " ")
		out, _ := in.ReadString('\n')
		a = strings.ReplaceAll(out, "\n", "")
		if allowEmpty && len(a) == 0 {
			if len(defaultValue) != 0 {
				return defaultValue[0]
			} else {
				return ""
			}
		}
	}

	return a
}

// user input with bool output
func inputYesNo(q string) bool {
	for {
		in := bufio.NewReader(os.Stdin)
		fmt.Print(q + " (Y/n)")
		out, _ := in.ReadString('\n')
		first := strings.ToLower(string(out[0]))
		outsan := strings.ReplaceAll(out, "\n", "")
		if first == "y" {
			return true
		}
		if first == "n" {
			return false
		}
		if len(outsan) == 0 {
			return true
		}

	}
}

// Makes the files
func createFiles(p *Project) {
	writeFile(p, "LICENSE", fillTempl(p, getFile("templates/licenses/"+p.License)))
	writeFile(p, "VERSION", p.Version)
	writeFile(p, "README.md", fillTempl(p, getFile("templates/readme/CLI.md")))
}

// Fills the template with project data
func fillTempl(p *Project, data string) string {
	t, err := template.New("Project").Parse(data)
	if err != nil {
		panic(err)
	}
	var s strings.Builder
	err = t.Execute(&s, p)
	if err != nil {
		panic(err)
	}
	return s.String()

}

// Gets the contents of the file from the embedFS
func getFile(f string) string {
	file, err := FStemplates.ReadFile(fmt.Sprintf("%v", f))
	if err != nil {
		panic(err)
	}
	return string(file)
}

// Writes the file
func writeFile(p *Project, filename string, data string) {
	path := filepath.Join(p.Directory, filename)
	fmt.Println("Writing.. ", path)
	file, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	_, err = file.WriteString(data)
	if err != nil {
		panic(err)
	}

}

// Lists all licenses
func list() {
	files, err := FStemplates.ReadDir("templates/licenses")

	if err != nil {
		panic(err)
	}
	fmt.Println("Licenses:")
	for _, i := range files {
		file, err := FStemplates.ReadFile(fmt.Sprintf("templates/licenses/%v", i.Name()))
		if err != nil {
			panic(err)
		}
		splitfile := strings.SplitN(string(file), "\n", 3)
		firstline := strings.TrimSpace(splitfile[0])
		secondline := strings.TrimSpace(splitfile[1])
		secondlineclean1 := strings.ReplaceAll(secondline, "=", "")
		var secondlinefinal string
		if len(secondlineclean1) != 0 {
			secondlinefinal = "(" + secondlineclean1 + ")"
		}
		fmt.Printf("%-30v - %v %v\n", i.Name(), firstline, secondlinefinal)
	}

}

// Views a specific license
func view(l string) {
	file, err := FStemplates.ReadFile(fmt.Sprintf("templates/licenses/%v", l))
	if err != nil {
		panic(err)
	}
	fmt.Println(string(file))
}
