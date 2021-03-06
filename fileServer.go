package main

import (
	"html/template"
	"log"
	"net/http"
	"path"
  "regexp"
  "io/ioutil"
  "fmt"
  "os"
  "time"

	"github.com/gorilla/mux"
  "github.com/google/uuid"
)

type fileServer interface {
	fileHandler(http.ResponseWriter, *http.Request)
}

type FileServer struct {
	nfs         fileBackend
	piInventory piManager
	diskManager *diskManager
}

type templatevars struct {
	PiId      string
	NfsServer string
	NfsRoot   string
}

func newFileServer(nfs fileBackend, inventory piManager, dm *diskManager) (fileServer, error) {
	return &FileServer{
    nfs:         nfs,
    piInventory: inventory,
    diskManager: dm,
	}, nil
}

func (f *FileServer) fileHandler(w http.ResponseWriter, r *http.Request) {
	urlvars := mux.Vars(r)
	filename := urlvars["filename"]
	piId := urlvars["piId"]

	//check if piId is allready registered. If not then register.
	pi, err := f.piInventory.GetPi(piId)
	if err != nil {
    log.Println("Pi not found in inventory. Putting a new one in the fridge.")
    pi = f.piInventory.NewPi(piId)
    err = pi.Save()
    if err != nil {
    	panic(err)
    }
	}

	//if filename == cmdline.txt then parse the template. else just serve the file
  bootLocation := f.nfs.GetBootRoot()+"/2018-04-18-raspbian-stretch-lite"
  if pi.Status != NOTINUSE && pi.Status != BOOTING {
	   bootLocation = pi.SourceBakeform.bootLocation
  }
	//strings.Replace(bootLocation, "/", "", 1) //remove the first /
	fullFilename := path.Join(bootLocation, filename)

  //log.Printf("%v requested", fullFilename)
	if filename == "cmdline.txt" {
    c := templatevars{
      NfsServer: f.nfs.GetNfsAddress(),
      NfsRoot:   f.nfs.GetNfsRoot()+"/481f6eed-275a-4456-9614-0d088dac4a41",
    }
    if pi.Status != NOTINUSE && pi.Status != BOOTING {
      c = templatevars{
      	NfsServer: f.nfs.GetNfsAddress(),
      	NfsRoot:   pi.Disks[0].Location,
      }
    }
    log.Printf("%v requested for: %v\n", filename, pi.Id)
    t, err := template.New("templatefile").ParseFiles(templatePath+"/cmdline.txt")
    if err != nil {
    	panic(err)
    }
    t.ExecuteTemplate(w, filename, c)
    return
	} else if filename == "config.txt" {
      // we want to make sure we are enabling uart on rpi3s so inject into config.txt
      // generate a new config.txt to serve temporarily from the original one without modifying it directly
      config_file, err := ioutil.ReadFile(fullFilename) // just pass the file name
    	if err != nil {
        fmt.Println(err)
    	}
      commentlines := regexp.MustCompile("(?m)[\r\n]+^#.*$")
      nocomments := commentlines.ReplaceAllString(string(config_file), "")
      pattern := regexp.MustCompile("enable_uart=[0-1]")
    	match := pattern.FindString(string(nocomments))
    	config_file_new := string(nocomments)
    	if len(match) == 0 {
    	config_file_new = string(nocomments) + "\nenable_uart=1"
    	}
    	config_file_save := pattern.ReplaceAllString(config_file_new, "enable_uart=1")
      random_uuid := uuid.New().String()
    	err = ioutil.WriteFile("/tmp/"+random_uuid, []byte(config_file_save), 0755)
      if err != nil {
        fmt.Println(err.Error())
    	}
      http.ServeFile(w, r, "/tmp/"+random_uuid)
      err = os.Remove("/tmp/"+random_uuid)
      if err != nil {
        fmt.Println(err.Error())
      }
  	  //return
  } else {
    http.ServeFile(w, r, fullFilename)
	  return
  }

  log.Printf("PiID: %v, PiStatus: %v", pi.Id, pi.Status)
  // check if the pi status is booting, if it is not in use and booting poll bushwood to see if it knows the serial number yet
  if pi.Status == NOTINUSE {
    pi.SetStatus(BOOTING)
    //Pi is not in inventory or not in use. Then don't serve files and power it off
    log.Printf("Pi %v came online but it's not in use. Powering it off\n", pi.Id)

    // poll bushwood
    var netClient = &http.Client{
      Timeout: time.Second * 10,
    }
    token := bushwoodToken
    req, _ := http.NewRequest("GET", bushwoodServer+"/api/v1/slots/pi/"+pi.Id, nil)
    req.Header.Add("apikey", token)
    timeout_count := 0
    //loop 4 times and break (30 seconds) timeout if no result
    for {
      if timeout_count == 6 {
        break
      }
      resp, _ := netClient.Do(req)
      log.Printf("%v%v%v", bushwoodServer,"/api/v1/slots/pi/",pi.Id)
      body, _ := ioutil.ReadAll(resp.Body)
      textBytes := []byte(body)
      resp.Body.Close()
      slotId := string(textBytes)[1 : len(string(textBytes))-1]
      //check if our response is empty, if it is, sleep, increment, and try again
      if string(slotId) == "" {
        time.Sleep(time.Second * 5)
        timeout_count++
      } else {
        break
      }
    }
    pi.SetStatus(NOTINUSE)
    log.Printf("PiID: %v, PiStatus: %v", pi.Id, pi.Status)
    err = pi.PowerOff()
    if err != nil {
      log.Println("A Pi just came online but I can't control its power state. Error:" + err.Error())
    }
    //w.WriteHeader(http.StatusNotFound)
    return
  }
}
