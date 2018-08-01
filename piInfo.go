package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
  "io/ioutil"
  "net/http"
  //"bytes"

  //"github.com/eclipse/paho.mqtt.golang"
)

type piStatus int

const (
	NOTINUSE  piStatus = 1
	INUSE     piStatus = 2
	PREPARING piStatus = 3
	BOOTING   piStatus = 4
)


type piList map[string]PiInfo

type PiInfo struct {
	db             *sql.DB
	Id             string     `json:"id"`
	Status         piStatus   `json:"status"`
	Disks          []*disk    `json:"disks,omitempty"`
	SourceBakeform *Bakeform  `json:"sourceBakeform,omitempty"`
	ppiPath        string
	ppiConfigPath  string
}

func (p *PiInfo) SetStatus(status piStatus) error {
	p.Status = status
	return p.Save()
}

func (p *PiInfo) Save() error {
	bakeformString := ""
	if p.SourceBakeform != nil {
		bakeformString = p.SourceBakeform.Name
	}

	_, err := p.db.Exec(fmt.Sprintf("insert into inventory(id, status, bakeform, diskIds) values('%v', %v, '%v', '%v')", p.Id, p.Status, bakeformString, ""))
	if err != nil {
		if err.Error() == "UNIQUE constraint failed: inventory.id" {
			var diskIds []string
			for _, diskStruct := range p.Disks {
				if diskStruct != nil {
					diskIds = append(diskIds, diskStruct.ID)
				}
			}

			diskIdsString := strings.Join(diskIds, ",")
			stmt := fmt.Sprintf("update inventory set status = %v, bakeform = '%v', diskIds = '%v' where id = '%v'", p.Status, bakeformString, diskIdsString, p.Id)
			_, err := p.db.Exec(stmt)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

func (p *PiInfo) Unbake(dm *diskManager) error {
	log.Printf("Unbaking pi with id: %v\n", p.Id)
	err := p.PowerOff()
	if err != nil {
		log.Println(err.Error())
		return err
	}

	//Get disk locations before deleting the pi
	disks := p.Disks

	//Set state to NOTINUSE and Store State
	p.Status = NOTINUSE
	p.SourceBakeform = nil
	err = p.Save()
	if err != nil {
		return err
	}

	//delete attached disks (including root)
	for _, d := range disks {
		if d == nil {
			continue
		}
		log.Println("Destroying disk: " + d.Location)
		err := dm.DestroyDisk(d.ID)
		if err != nil {
			log.Println(err.Error())
		}
	}

	return err
}

func (p *PiInfo) doPpiAction(action string) error {
	if action != "on" && action != "off" && action != "hardoff"{
		return fmt.Errorf("action %v not supported", action)
	}
	params := ppiParams{
		PiId:   p.Id,
		Action: action,
	}
  var netClient = &http.Client{
    Timeout: time.Second * 10,
  }
  token := bushwoodToken
  req, _ := http.NewRequest("GET", bushwoodServer+"/api/v1/slots/pi/"+params.PiId, nil)
  req.Header.Add("apikey", token)
  resp, _ := netClient.Do(req)
  log.Printf("%v%v%v", bushwood,"/api/v1/slots/pi/",params.PiId)
  defer resp.Body.Close()
  body, _ := ioutil.ReadAll(resp.Body)
  textBytes := []byte(body)
  slotId := string(textBytes)[1 : len(string(textBytes))-1]
  if string(slotId) == "" {
    log.Printf("No matching slot for %v", params.PiId)
  } else {
    log.Printf("%v %v %v", params.Action, params.PiId, string(slotId))
    req, _ := http.NewRequest("POST", bushwoodServer+"/api/v1/slots/"+string(slotId)+"/power/"+params.Action, nil)
    req.Header.Add("apikey", token)
    resp, _ := netClient.Do(req)
    body, _ := ioutil.ReadAll(resp.Body)
    log.Printf(string(body))
    defer resp.Body.Close()
  }
	return nil
}

func (p *PiInfo) PowerOn() error {
	return p.doPpiAction("on")
}

func (p *PiInfo) PowerOff() error {
	return p.doPpiAction("hardoff") //hard power off
}

func (p *PiInfo) PowerCycle() error {
	err := p.doPpiAction("hardoff")
	if err != nil {
		return err
	}

	time.Sleep(1 * time.Second)

	return p.doPpiAction("on")
}

func (p *PiInfo) AttachDisk(dsk *disk) error {
	//Check if disk is already attached. Early return if so
	for _, disk := range p.Disks {
		if disk == dsk {
			log.Printf("AttachDisk: disk %v already attached", dsk.ID)
			return nil
		}
	}

	//not attached? attach now and save.
	p.Disks = append(p.Disks, dsk)
	log.Printf("AttachDisk: Attached %v", dsk.ID)
	return p.Save()
}

func (p *PiInfo) DetachDisk(dsk *disk) error {
	for i, d := range p.Disks {
		//find the disk in the array
		if d.ID == dsk.ID {
			if i == 0 { //do not try to detach system disk
				return fmt.Errorf("Cannot unassociate boot volume")
			}
			//delete disk from array
			p.Disks = append(p.Disks[:i], p.Disks[i+1:]...)
			log.Printf("DetachDisk: disk %v detached", dsk.ID)
			return p.Save()
		}
	}
	//if loop exits we didn't find the disk in the Pi
	return fmt.Errorf("Disk with id %v is not attached to pi with id %v", dsk.ID, p.Id)
}
