package main

import (
    "flag"
    "fmt"
    "github.com/larspensjo/config"
    "log"
    "time"
    "runtime"
    "encoding/json"
    "os/exec"
    "strings"
    "os"
    "path/filepath"
    "io/ioutil"
    "strconv"
)

var (
    configFile = flag.String("configfile", "config.ini", "General configuration file")
    prepareData = flag.String("predata", "yes or no", "determion fill in disk with data")
)

var TOPIC = make(map[string]string)

type Fio struct{
     bs string `json:"bs"`
     rw string `json:"rw"`
     ioengine string `json:"ioengine"`
     size string `json:"size"`
     filename string `json:"filename"`
     numjobs string `json:"numjobs"`
     runtime string `json:"runtime"`
     iodepth string `json:"iodepth"`
     rwmixread string `json:"rwmixread"`
     case_name string `json:"case_name"`
     sleep_time string `json:"sleep_time"`
     pre_exec_time string `json:"pre_exec_time"`
     vm_ip string `json:"vm_ip"`
     sum_ip string `json:"sum_ip"`
     host_ip string `json: "host_ip"`
}

type Test interface{
     ExecShell()
     GetExecTime()
     GenTask()
     RunTask()
     ParseResult()
     PreData()
}

func (fio *Fio) ExecShell(shell_script string){
    cmd := exec.Command("sh", "-c", shell_script)
    if _, err := cmd.Output(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}

func (fio *Fio) GenTask(mode string, rw_type string, size string, numjobs string,
    ioeng string, runtime string, filename string, rwmixread string) string {
     return fmt.Sprintf("%s %s %s %s %s %s %s %s",mode, rw_type, size, numjobs,
                        ioeng, runtime, filename, rwmixread)
}
func (fio *Fio) GetExecTime(dates time, second int) string {
    timeUnix := dates.Uninx() + second
    formatTimeStr:=time.Unix(timeUnix,0).Format("2006-01-02 15:04:05")
    return fmt.Sprintf("%s %s %s", formatTimeStr.Minute(), formatTimeStr.Hour(), formatTimeStr.Day())
}

func (fio *Fio) GenTask() []string {
    tasks := make([]string , 20)
    rw_type := strings.Split(fio.rw, " ")
    bs_type := strings.Split(fio.bs, " ")
    for i, v := range rw_type {
        for j, b := range bs_type {
            append(tasks, fio.GenTask(v, b, fio.size, fio.numjobs, fio.ioengine, fio.runtime,
                                    fio.filename, fio.rwmixread))
        }
    }
    return tasks
}

func (fio *Fio) RunTask(shell_script string, dates time, task string, ips []string, sum_ip string, data_dir string){
    cmds := fmt.Sprintf("%s %s %s", shell_script, dates, task)
    fio.ExecShell(cmds)
    for ip := range ips{
        fio_log := fmt.Sprintf("%s/%s_fio", data_dir, ip)
        _, err := os.Stat(fio_log)
	    if err != nil {
            os.Mkdir(fio_log)
	    }
        task_split := strings.Split(task, " ")
        task_split[1]=fmt.Sprintf("%sk", task_list[1])
        fio_log_name = fmt.Spintf("%s_%s.json", strings.Join(task_list[:-2], "_"), ip)
        cmds := fmt.Sprintf("start_fio.sh %s %s %s",ip, sum_ip, fio_log_name)
        fio.ExecShell(cmds)
    }
}

func (fio *Fio) ParseResult(){
    var result map[string]interface{}
    var host_result map[string]interface{}
    var report_result map[string]interface{}
    parent_path := "/root/result_fio_report"
    num := 0
    err := filepath.Walk(parent_path, func(path string, fs os.FileInfo, err error) error {
        if fs.IsDir() {
            num ++
            log_dir := filepath.Join(parent_path, fs)
            err := filepath.Walk(log_dir, func(path string, f os.FileInfo, err error) error{
                fmt.Println(f)
                files_name_list := strings.Split(f)
                key := files_name_list[1]
                mode := file_name_list[0]
                new_key := fmt.Sprintf("%s_%s", key, mode)
                var modes []string
                if mode == "randwrite"{
                    append(modes, "write")
                }else if mode == "randread"{
                    append(modes, "read")
                }else if mode == "read"{
                    append(modes, "read")
                }else if mode == "write"{
                    append(modes, "write")
                }else if mode == "randrw"{
                    append(modes ,"read")
                    append(modes, "write")
                }
                files_name := filepath.Join(parent_path, log_dir, f)
                var jsonObject = map[string]string{}
                bytes, err := ioutil.ReadFile(files_name)
                if err != nil {
                    fmt.Println("ReadFile: ", err.Error())
                    return nil, err
                }
                if err := json.Unmarshal(bytes, &jsonObject); err != nil {
                    fmt.Println("Unmarshal: ", err.Error())
                    return nil, err
                }
                for m := range modes{
                    content := jsonObject["jobs"][0][m]
                    fbw, err := strconv.ParseFloat(content["bw"], 64)
                    bw := fbw / 1024
                    iops := content["iops"]
                    flat, err := strconv.ParseFloat(content["lat"]["mean"], 64)
                    lat := flat / 1000

                    host_result_json, ok := host_result[fs]
                    if !ok{
                        var host_result_json map[string] interface{}
                        host_result_json = make(map[string]interface{})
                        host_result[fs] = host_result_json
                    }
                    key_item_result, ok := host_result_dict[new_key]
                    if !ok{
                        var key_item_result map[string] interface{}
                        key_item_result= make(map[string]interface{})
                        host_result_dict[new_key] = key_item_result
                    }
                    item_m_result, ok := key_item_result[m]
                    if !ok{
                        var item_m_result map[string] interface{}
                        item_m_result = make(map[string]interface{})
                        key_item_result[m] = item_m_result
                    }
                    sbw := strconv.FormatFloat(bw, 'f', 6, 64)
                    item_m_result["bw"] = sbw + " MB/s"
                    item_m_result["iops"] = iops
                    slat := strconv.FormatFloat(lat, 'f', 6, 64)
                    item_m_result["lat"] = slat + " ms"

                    all_result_dict, ok := result[new_key]
                    if !ok{
                        var all_result_dict map[string]interface{}
                        all_result_dict = make(map[string]interface{})
                        result[new_key] = all_result_dict
                    }
                    item_all_m_dict ,ok :=  all_result_dict[m]
                    if !ok{
                        var item_all_m_dict map[string]interface{}
                        item_all_m_dict = make(map[string]interface{})
                        all_result_dict[new_key] = item_all_m_dict
                    }
                    nbw, ok := item_all_m_dict["bw"]
                    if !ok{
                        nbw := 0
                    }
                    item_all_m_dict["bw"] = nbw + bw
                    niops, ok := item_all_m_dict["iops"]
                    if !ok{
                        niops := 0
                    }
                    item_all_m_dict["iops"] = niops + iops
                    nlat, ok := item_all_m_dict["lat"]
                    if !ok{
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

    for key, value := range result{
        for keys, values := range value{
            sbw := strconv.FormatFloat(values["bw"], 'f', 6, 64)
            values["bw"] = sbw + "MB/s"
            slat := strconv.FormatFloat(values["lat"], 'f', 6, 64)
            values["lat"] = (slat / num) + " ms"

        }
    }
        
    report_result = make(map[string]interface{})
    report_result["summarry"] = result
    report_result["host_detail"] = host_result

    datetimes := fmt.Sprintf("%s%s%s%s", time.Now().Year(),time.Now().Month(), time.Now().Day(), time.Now().Hour())
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

func conf_args(){
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
     section ,err = cfg.SectionOptions("host")
     if err == nil {
         for _, s := range section {
              options, err := cfg.String("host", s)
                  if err == nil {
                      TOPIC[s] = options
                  }
         }
     }
}

func main(){
    runtime.GOMAXPROCS(runtime.NumCPU())
    flag.Parse()
    conf_args()
    fmt.Println(TOPIC)
    var fio *Fio = &Fio{TOPIC["bs"], TOPIC["rw"], TOPIC["ioengine"], TOPIC["size"], TOPIC["filename"], TOPIC["numjobs"], TOPIC["runtime"], TOPIC["iodepth"],TOPIC["rwmixread"],TOPIC["case_name"], TOPIC["sleep_time"], TOPIC["pre_exec_time"], TOPIC["vm_ip"], TOPIC["sum_ip"], TOPIC["host_ip"]}
    
    
}
