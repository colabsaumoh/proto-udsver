
package os
import (
	"bufio"
	"errors"
	"os"
	"regexp"
	"log"
	"strconv"
	"strings"
)

func GetContainerId(pathPrefix string, pid int32) (cid string, err error) {
	path := pathPrefix + "/proc/" + strconv.Itoa(int(pid)) + "/cgroup"
	re := regexp.MustCompile("^1:name")
	file, err := os.Open(path)
	if err != nil {
		errS := "Not able to open proc file " + path + " (" + err.Error() + ")"
		return "", errors.New(errS)
	}
	defer file.Close()

	var rstr string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		rstr = scanner.Text()
		r := re.FindString(rstr)
		if r != "" {
			break
		}
	}
	if rstr == "" {
		return "", errors.New("Not able to find the container id")
	}

	vals := strings.Split(rstr, "/")
	if vals[0] == rstr {
		log.Println(vals)
		return "", errors.New("The cgroups does not contain CID")
	}

	return vals[len(vals)-1], nil
}
