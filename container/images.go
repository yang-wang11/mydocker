package container

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"
)

func ListImages(output bool) []string {

	var ims []string

	// list the files from images folder
	images, err := ioutil.ReadDir(ImageBaseFolder);
	if err != nil {
		log.Errorf("list images folder failed.")
		return ims
	}

	// print images
	for _, image := range images {
		ims = append(ims, image.Name())
		if output {
			fmt.Println(image.Name())
		}
	}

  return ims

}


func CheckImage(imageName string) bool {

	matched := false

	supportedImages := ListImages(false)

	for _, img := range supportedImages {
		if img == imageName {
			matched = true
			break
		}
	}

	if !matched {
		return false
	} else {
		return true
	}

}

func RemoveImages(imageName string) bool {

	if !CheckImage(imageName) {
		return false
	}

	imagePath := path.Join(ImageBaseFolder, imageName)

	if err := os.Remove(imagePath); err != nil {
		log.Errorf("delete image %s failed, %v", imageName, err)
		return false
	} else {
		log.Debugf("delete image %s successfully", imageName)
		return true
	}

}