package main

import (
	_ "encoding/json"
	"flag"
	"fmt"
	_ "io/ioutil"
	"log"
	"os"
	"os/exec"
	_ "path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
	"github.com/larspensjo/config"
)

var (
	configFile  = flag.String("configfile", "config.ini", "General configuration file")
	prepareData = flag.String("predata", "yes or no", "determion fill in disk with data")
)

var TOPIC = make(map[string]string)

type Fio struct {
	bs            string `json:"bs"`
	rw            string `json:"rw"`
	ioengine      string `json:"ioengine"`
	size          string `json:"size"`
	filename      string `json:"filename"`
	numjobs       string `json:"numjobs"`
	runtime       string `json:"runtime"`
	iodepth       string `json:"iodepth"`
	rwmixread     string `json:"rwmixread"`
	case_name     string `json:"case_name"`
	sleep_time    string `json:"sleep_time"`
	pre_exec_time string `json:"pre_exec_time"`
	vm_ip         string `json:"vm_ip"`
	sum_ip        string `json:"sum_ip"`
	host_ip       string `json: "host_ip"`
}

type Test interface {
	ExecShell()
	GetExecTime() string
	GenTask() string
	RunTask()
	ParseResult()
	PreData()
}

func (fio Fio) ExecShell(shell_script string) {
	cmd := exec.Command("/bin/sh", "-c", shell_script)
	if  _, err := cmd.Output(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func (fio Fio) GenTask(mode string, rw_type string, size string, numjobs string,
	ioeng string, runtime string, filename string, rwmixread string) string {
	return fmt.Sprintf("%s %s %s %s %s %s %s %s", mode, rw_type, size, numjobs,
		ioeng, runtime, filename, rwmixread)
}
func (fio Fio) GetExecTime(dates time.Time, second int64) string {
	timeUnix := dates.Unix() + second
        formatTimeStr :=time.Unix(timeUnix,0).Format("2006-01-02 15:04:05")
	formatTime , _:= time.Parse("2006-01-02 15:04:05", formatTimeStr)
	return fmt.Sprintf("%d %d %d", formatTime.Minute(), formatTime.Hour(), formatTime.Day())
}

func (fio Fio) GenTaskList() []string {
	tasks := make([]string, 20)
	rw_type := strings.Split(fio.rw, " ")
	bs_type := strings.Split(fio.bs, " ")
	for _, v := range rw_type {
	    for _, b := range bs_type {
	        tasks = append(tasks, fio.GenTask(v, b, fio.size, fio.numjobs, fio.ioengine, fio.runtime,
	                      fio.filename, fio.rwmixread))
		}
	}
	return tasks
}

func (fio Fio) PreData(pre_task string , vm_ips []string, sleep_second int, pre_runtime int){
    day_hour_minute := fio.GetExecTime(time.Now(), 120)
    fio.RunTask("./gen_crontab.sh", day_hour_minute, pre_task, vm_ips, "", "")
    sleep_time :=  sleep_second + pre_runtime + 120
    time.Sleep(time.Duration(sleep_time) * time.Second)
    for ip := range vm_ips{
        kill_cmd := fmt.Sprintf("kill_fio.sh '%s'", ip)
        fio.ExecShell(kill_cmd)
}
}

func (fio Fio) RunTask(shell_script string, dates string, task string, ips []string, sum_ip string, data_dir string) {
	cmds := fmt.Sprintf("%s \"%s\" \"%s\"", shell_script, dates, task)
	fio.ExecShell(cmds)
        fmt.Println(cmds)
	for ip := range ips {
		fio_log := fmt.Sprintf("%s/%s_fio", data_dir, ip)
		_, err := os.Stat(fio_log)
		if err != nil {
			os.Mkdir(fio_log, 0777)
		}
		task_split := strings.Split(task, " ")
		task_bs := fmt.Sprintf("%sk", task_split[1])
		fio_log_name := fmt.Sprintf("%s_%s_%s.json",task_bs, strings.Join(task_split[:7], "_"), ip)
		cmds := fmt.Sprintf("./start_fio.sh %s %s %s", ip, sum_ip, fio_log_name)
		fio.ExecShell(cmds)
	}
}
/*
func (fio Fio) ParseResult() {
	var result map[string]interface{}
	var host_result map[string]interface{}
	var report_result map[string]interface{}
	parent_path := "/root/result_fio_report"
	num := 0
	err := filepath.Walk(parent_path, func(path string, fs os.FileInfo, err error) error {
		if fs.IsDir() {
			num++
			log_dir := filepath.Join(parent_path, path)
			err := filepath.Walk(log_dir, func(pathf string, f os.FileInfo, err error) error {
				fmt.Println(path)
				files_name_list := strings.Split(path, "_")
				key := files_name_list[1]
				mode := files_name_list[0]
				new_key := fmt.Sprintf("%s_%s", key, mode)
				var modes []string
				if (mode == "randwrite") {
				    modes = append(modes, "write")
				} else if (mode == "randread") {
				    modes = append(modes, "read")
				} else if (mode == "read") {
				    modes = append(modes, "read")
				} else if (mode == "write") {
				    modes = append(modes, "write")
				} else if (mode == "randrw") {
				    modes = append(modes, "read")
				    modes = append(modes, "write")
				}
				files_name := filepath.Join(parent_path, log_dir, pathf)
				jsonObject := make(map[string]interface{})
				bytes, err := ioutil.ReadFile(files_name)
				if err != nil {
					fmt.Println("ReadFile: ", err.Error())
					return err
				}
				if err := json.Unmarshal(bytes, &jsonObject); err != nil {
					fmt.Println("Unmarshal: ", err.Error())
					return err
				}
				for m := range modes {
					content := jsonObject["jobs"][0][m]
					fbw, err := strconv.ParseFloat(content["bw"], 64)
					bw := fbw / 1024
					iops := content["iops"]
					flat, err := strconv.ParseFloat(content["lat"]["mean"], 64)
					lat := flat / 1000

					host_result_json, ok := host_result[path]
					if !ok {
						var host_result_json map[string]interface{}
						host_result_json = make(map[string]interface{})
						host_result[path] = host_result_json
					}
					key_item_result, ok := host_result_dict[new_key]
					if !ok {
						var key_item_result map[string]interface{}
						key_item_result = make(map[string]interface{})
						host_result_dict[new_key] = key_item_result
					}
					item_m_result, ok := key_item_result[m]
					if !ok {
						var item_m_result map[string]interface{}
						item_m_result = make(map[string]interface{})
						key_item_result[m] = item_m_result
					}
					sbw := strconv.FormatFloat(bw, 'f', 6, 64)
					item_m_result["bw"] = sbw + " MB/s"
					item_m_result["iops"] = iops
					slat := strconv.FormatFloat(lat, 'f', 6, 64)
					item_m_result["lat"] = slat + " ms"

					all_result_dict, ok := result[new_key]
					if !ok {
						var all_result_dict map[string]interface{}
						all_result_dict = make(map[string]interface{})
						result[new_key] = all_result_dict
					}
					item_all_m_dict, ok := all_result_dict[m]
					if !ok {
						var item_all_m_dict map[string]interface{}
						item_all_m_dict = make(map[string]interface{})
						all_result_dict[new_key] = item_all_m_dict
					}
					nbw, ok := item_all_m_dict["bw"]
					if !ok {
						nbw := 0
					}
					item_all_m_dict["bw"] = nbw + bw
					niops, ok := item_all_m_dict["iops"]
					if !ok {
						niops := 0
					}
					item_all_m_dict["iops"] = niops + iops
					nlat, ok := item_all_m_dict["lat"]
					if !ok {
						nlat := 0
					}
					item_all_m_dict["lat"] = nlat + lat

				}
			})
		}
		return nil
	})
	if err != nil {
		fmt.Printf("filepath.Walk() returned %v\n", err)
	}

	for key, value := range result {
		for keys, values := range value {
			sbw := strconv.FormatFloat(values["bw"], 'f', 6, 64)
			values["bw"] = sbw + "MB/s"
			slat := strconv.FormatFloat(values["lat"], 'f', 6, 64)
			values["lat"] = (slat / num) + " ms"

		}
	}

	report_result = make(map[string]interface{})
	report_result["summarry"] = result
	report_result["host_detail"] = host_result

	datetimes := fmt.Sprintf("%s%s%s%s", time.Now().Year(), time.Now().Month(), time.Now().Day(), time.Now().Hour())
	report_name := fmt.Sprintf("fio-test-report-%s-%s.json", datetimes, fio.case_name)
	report_file_path = filepath.Join("/root", report_name)
	fp, err := os.OpenFile(report_file_path, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err)
	}
	defer fp.Close()
	enc := json.NewEncoder(fp)
	enc.Encode(resport_result)
}
*/
func conf_args() {
	cfg, err := config.ReadDefault(*configFile)
	if err != nil {
		log.Fatalf("Fail to find", *configFile, err)
	}
	//set config file std End
	//Initialized topic from the configuration
	section, err := cfg.SectionOptions("default")
	if err == nil {
		for _, v := range section {
			options, err := cfg.String("default", v)
			if err == nil {
				TOPIC[v] = options
			}
		}
	}
	section, err = cfg.SectionOptions("host")
	if err == nil {
		for _, s := range section {
			options, err := cfg.String("host", s)
			if err == nil {
				TOPIC[s] = options
			}
		}
	}
}

func main() {
    runtime.GOMAXPROCS(runtime.NumCPU())
    flag.Parse()
    conf_args()
    fmt.Println(TOPIC)
    var fio Fio
    fio = Fio{TOPIC["bs"], TOPIC["rw"], TOPIC["ioengine"], TOPIC["size"], TOPIC["filename"], TOPIC["numjobs"], TOPIC["runtime"],
                    TOPIC["iodepth"], TOPIC["rwmixread"], TOPIC["case_name"], TOPIC["sleep_time"], TOPIC["pre_exec_time"],
                    TOPIC["vm_ip"], TOPIC["sum_ip"], TOPIC["host_ip"]}
    fio_log_dir := "/root/result_fio_report"
    shell_cmd := fmt.Sprintf("rm -rf %s", fio_log_dir)
    fio.ExecShell(shell_cmd)
    os.Mkdir(fio_log_dir,0777)
    tasks := fio.GenTaskList()
    sleep_second, _ := strconv.Atoi(fio.sleep_time)
    //pre_runtime, err := strconv.Atoi(fio.pre_exec_time)
    fio_runtime, _ :=  strconv.Atoi(fio.runtime)
    fio_vm_ips := strings.Split(fio.vm_ip, " ")
    host_ips := strings.Split(fio.host_ip, " ")
    mode_list :=[]string{"randwrite","randread","randrw"}
    for _, hip := range host_ips{
	clear_cmd := fmt.Sprintf("scp fio/clear_caches.sh root@%s:~/", hip)
	fio.ExecShell(clear_cmd)
        fmt.Println(clear_cmd)
    }

    if *prepareData == "yes"{
	task := fio.GenTask("randrw", "1024", fio.size, "2",
			"libaio", "1000", fio.filename, "50")
        fmt.Println(task)
	fio.PreData(task, fio_vm_ips, sleep_second, 1000)
    }
    fmt.Println("sfasdfasdf") 
    for _, tsk := range tasks{
	for hip :=range  host_ips{
            drop_cmd := fmt.Sprintf("drop_caches.sh '%s'", hip)
            fio.ExecShell(drop_cmd)
        }
        current_task := strings.Split(tsk, " ")
        bs_size, _ := strconv.Atoi(current_task[1])
        fmt.Println(bs_size)
        if (bs_size >= 1024){
	    tag := false
	    for _, m := range mode_list{
	  	if m == current_task[1]{
	           tag = true
		}
	    }
	    if tag{
		continue
	   }
        } 
        pre_task := fio.GenTask(current_task[0], current_task[1], fio.size, fio.numjobs,
                                fio.ioengine, fio.pre_exec_time, fio.filename, fio.rwmixread)

        fmt.Println("current task:  %s", tsk)
        pre_runtime, _ := strconv.Atoi(fio.pre_exec_time)
        fio.PreData(pre_task, fio_vm_ips, sleep_second, pre_runtime)

        minute := fio.GetExecTime(time.Now(), 120)
        fio.RunTask("gen_crontab.sh", minute, tsk, fio_vm_ips, fio.sum_ip, fio_log_dir)
        task_sleep := fio_runtime + sleep_second + 120
        time.Sleep(time.Duration(task_sleep) * time.Second)
	for _, ip := range  fio_vm_ips{
	    kill_cmd := fmt.Sprintf("kill_fio.sh '%s'", ip)
	    fio.ExecShell(kill_cmd)
	    get_cmd := fmt.Sprintf("get_result.sh '%s'", ip)
	    fio.ExecShell(get_cmd)
	    drop_cmd := fmt.Sprintf("drop_caches.sh '%s'", ip)
            fio.ExecShell(drop_cmd)
	}
	//fio.ParseResult(fio.case_name)
  }
}
