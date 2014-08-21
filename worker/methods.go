package worker

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/iron-io/iron_go/api"
)

type Schedule struct {
	CodeName       string         `json:"code_name"`
	Delay          *time.Duration `json:"delay"`
	EndAt          *time.Time     `json:"end_at"`
	MaxConcurrency *int           `json:"max_concurrency"`
	Name           string         `json:"name"`
	Payload        string         `json:"payload"`
	Priority       *int           `json:"priority"`
	RunEvery       *int           `json:"run_every"`
	RunTimes       *int           `json:"run_times"`
	StartAt        *time.Time     `json:"start_at"`
}

type ScheduleInfo struct {
	CodeName       string    `json:"code_name"`
	CreatedAt      time.Time `json:"created_at"`
	EndAt          time.Time `json:"end_at"`
	Id             string    `json:"id"`
	LastRunTime    time.Time `json:"last_run_time"`
	MaxConcurrency int       `json:"max_concurrency"`
	Msg            string    `json:"msg"`
	NextStart      time.Time `json:"next_start"`
	ProjectId      string    `json:"project_id"`
	RunCount       int       `json:"run_count"`
	RunTimes       int       `json:"run_times"`
	StartAt        time.Time `json:"start_at"`
	Status         string    `json:"status"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Task struct {
	CodeName string         `json:"code_name"`
	Payload  string         `json:"payload"`
	Priority int            `json:"priority"`
	Timeout  *time.Duration `json:"timeout"`
	Delay    *time.Duration `json:"delay"`
}

type TaskInfo struct {
	CodeHistoryId string    `json:"code_history_id"`
	CodeId        string    `json:"code_id"`
	CodeName      string    `json:"code_name"`
	CodeRev       string    `json:"code_rev"`
	Id            string    `json:"id"`
	Payload       string    `json:"payload"`
	ProjectId     string    `json:"project_id"`
	Status        string    `json:"status"`
	Msg           string    `json:"msg,omitempty"`
	Duration      int       `json:"duration"`
	RunTimes      int       `json:"run_times"`
	Timeout       int       `json:"timeout"`
	Percent       int       `json:"percent,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
}

type CodeSource map[string][]byte // map[pathInZip]code

type Code struct {
	Name           string        `json:"name"`
	Runtime        string        `json:"runtime"`
	FileName       string        `json:"file_name"`
	Config         string        `json:"config,omitempty"`
	MaxConcurrency int           `json:"max_concurrency,omitempty"`
	Retries        int           `json:"retries,omitempty"`
	RetriesDelay   time.Duration `json:"-"`
	Source         CodeSource    `json:"-"`
}

type CodeInfo struct {
	Id              string    `json:"id"`
	LatestChecksum  string    `json:"latest_checksum"`
	LatestHistoryId string    `json:"latest_history_id"`
	Name            string    `json:"name"`
	ProjectId       string    `json:"project_id"`
	Runtime         string    `json:"runtime"`
	Rev             int       `json:"rev"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	LatestChange    time.Time `json:"latest_change"`
}

// CodePackageList lists code packages.
//
// The page argument decides the page of code packages you want to retrieve, starting from 0, maximum is 100.
//
// The perPage argument determines the number of code packages to return. Note
// this is a maximum value, so there may be fewer packages returned if there
// aren’t enough results. If this is < 1, 1 will be the default. Maximum is 100.
func (w *Worker) CodePackageList(page, perPage int) (codes []CodeInfo, err error) {
	out := map[string][]CodeInfo{}

	err = w.codes().
		QueryAdd("page", "%d", page).
		QueryAdd("per_page", "%d", perPage).
		Req("GET", nil, &out)
	if err != nil {
		return
	}

	return out["codes"], nil
}

// CodePackageUpload uploads a code package
func (w *Worker) CodePackageUpload(code Code) (id string, err error) {
	client := http.Client{}

	body := &bytes.Buffer{}
	mWriter := multipart.NewWriter(body)

	// write meta-data
	mMetaWriter, err := mWriter.CreateFormField("data")
	if err != nil {
		return
	}
	jEncoder := json.NewEncoder(mMetaWriter)
	err = jEncoder.Encode(map[string]interface{}{
		"name":            code.Name,
		"runtime":         code.Runtime,
		"file_name":       code.FileName,
		"config":          code.Config,
		"max_concurrency": code.MaxConcurrency,
		"retries":         code.Retries,
		"retries_delay":   code.RetriesDelay.Seconds(),
	})
	if err != nil {
		return
	}

	// write the zip
	mFileWriter, err := mWriter.CreateFormFile("file", "worker.zip")
	if err != nil {
		return
	}
	zWriter := zip.NewWriter(mFileWriter)

	for sourcePath, sourceText := range code.Source {
		fWriter, err := zWriter.Create(sourcePath)
		if err != nil {
			return "", err
		}
		fWriter.Write([]byte(sourceText))
	}

	zWriter.Close()

	// done with multipart
	mWriter.Close()

	req, err := http.NewRequest("POST", w.codes().URL.String(), body)
	if err != nil {
		return
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Encoding", "gzip/deflate")
	req.Header.Set("Authorization", "OAuth "+w.Settings.Token)
	req.Header.Set("Content-Type", mWriter.FormDataContentType())
	req.Header.Set("User-Agent", w.Settings.UserAgent)

	// dumpRequest(req) NOTE: never do this here, it breaks stuff
	response, err := client.Do(req)
	if err != nil {
		return
	}
	if err = api.ResponseAsError(response); err != nil {
		return
	}

	// dumpResponse(response)

	data := struct {
		Id         string `json:"id"`
		Msg        string `json:"msg"`
		StatusCode int    `json:"status_code"`
	}{}
	err = json.NewDecoder(response.Body).Decode(&data)
	if err != nil {
		return
	}

	return data.Id, err
}

// CodePackageInfo gets info about a code package
func (w *Worker) CodePackageInfo(codeId string) (code CodeInfo, err error) {
	out := CodeInfo{}
	err = w.codes(codeId).Req("GET", nil, &out)
	return out, err
}

// CodePackageDelete deletes a code package
func (w *Worker) CodePackageDelete(codeId string) (err error) {
	return w.codes(codeId).Req("DELETE", nil, nil)
}

// CodePackageDownload downloads a code package
func (w *Worker) CodePackageDownload(codeId string) (code Code, err error) {
	out := Code{}
	err = w.codes(codeId, "download").Req("GET", nil, &out)
	return out, err
}

// CodePackageRevisions lists the revisions of a code pacakge
func (w *Worker) CodePackageRevisions(codeId string) (code Code, err error) {
	out := Code{}
	err = w.codes(codeId, "revisions").Req("GET", nil, &out)
	return out, err
}

func (w *Worker) TaskList() (tasks []TaskInfo, err error) {
	out := map[string][]TaskInfo{}
	err = w.tasks().Req("GET", nil, &out)
	if err != nil {
		return
	}
	return out["tasks"], nil
}

// TaskQueue queues a task
func (w *Worker) TaskQueue(tasks ...Task) (taskIds []string, err error) {
	outTasks := make([]map[string]interface{}, 0, len(tasks))

	for _, task := range tasks {
		thisTask := map[string]interface{}{
			"code_name": task.CodeName,
			"payload":   task.Payload,
			"priority":  task.Priority,
		}
		if task.Timeout != nil {
			thisTask["timeout"] = (*task.Timeout).Seconds()
		}
		if task.Delay != nil {
			thisTask["delay"] = (*task.Delay).Seconds()
		}

		outTasks = append(outTasks, thisTask)
	}

	in := map[string][]map[string]interface{}{"tasks": outTasks}
	out := struct {
		Tasks []struct {
			Id string `json:"id"`
		} `json:"tasks"`
		Msg string `json:"msg"`
	}{}

	err = w.tasks().Req("POST", &in, &out)
	if err != nil {
		return
	}

	taskIds = make([]string, 0, len(out.Tasks))
	for _, task := range out.Tasks {
		taskIds = append(taskIds, task.Id)
	}

	return
}

// TaskInfo gives info about a given task
func (w *Worker) TaskInfo(taskId string) (task TaskInfo, err error) {
	out := TaskInfo{}
	err = w.tasks(taskId).Req("GET", nil, &out)
	return out, err
}

func (w *Worker) TaskLog(taskId string) (log []byte, err error) {
	response, err := w.tasks(taskId, "log").Request("GET", nil)
	if err != nil {
		return
	}

	log, err = ioutil.ReadAll(response.Body)
	return
}

// TaskCancel cancels a Task
func (w *Worker) TaskCancel(taskId string) (err error) {
	_, err = w.tasks(taskId, "cancel").Request("POST", nil)
	return err
}

// TaskProgress sets a Task's Progress
func (w *Worker) TaskProgress(taskId string, progress int, msg string) (err error) {
	payload := map[string]interface{}{
		"msg":     msg,
		"percent": progress,
	}

	err = w.tasks(taskId, "progress").Req("POST", payload, nil)
	return
}

// TaskQueueWebhook queues a Task from a Webhook
func (w *Worker) TaskQueueWebhook() (err error) { return }

// ScheduleList lists Scheduled Tasks
func (w *Worker) ScheduleList() (schedules []ScheduleInfo, err error) {
	out := map[string][]ScheduleInfo{}
	err = w.schedules().Req("GET", nil, &out)
	if err != nil {
		return
	}
	return out["schedules"], nil
}

// Schedule a Task
func (w *Worker) Schedule(schedules ...Schedule) (scheduleIds []string, err error) {
	outSchedules := make([]map[string]interface{}, 0, len(schedules))

	for _, schedule := range schedules {
		sm := map[string]interface{}{
			"code_name": schedule.CodeName,
			"name":      schedule.Name,
			"payload":   schedule.Payload,
		}
		if schedule.Delay != nil {
			sm["delay"] = (*schedule.Delay).Seconds()
		}
		if schedule.EndAt != nil {
			sm["end_at"] = *schedule.EndAt
		}
		if schedule.MaxConcurrency != nil {
			sm["max_concurrency"] = *schedule.MaxConcurrency
		}
		if schedule.Priority != nil {
			sm["priority"] = *schedule.Priority
		}
		if schedule.RunEvery != nil {
			sm["run_every"] = *schedule.RunEvery
		}
		if schedule.RunTimes != nil {
			sm["run_times"] = *schedule.RunTimes
		}
		if schedule.StartAt != nil {
			sm["start_at"] = *schedule.StartAt
		}
		outSchedules = append(outSchedules, sm)
	}

	in := map[string][]map[string]interface{}{"schedules": outSchedules}
	out := struct {
		Schedules []struct {
			Id string `json:"id"`
		} `json:"schedules"`
		Msg string `json:"msg"`
	}{}

	err = w.schedules().Req("POST", &in, &out)
	if err != nil {
		return
	}

	scheduleIds = make([]string, 0, len(out.Schedules))

	for _, schedule := range out.Schedules {
		scheduleIds = append(scheduleIds, schedule.Id)
	}

	return
}

// ScheduleInfo gets info about a scheduled task
func (w *Worker) ScheduleInfo(scheduleId string) (info ScheduleInfo, err error) {
	info = ScheduleInfo{}
	err = w.schedules(scheduleId).Req("GET", nil, &info)
	return info, nil
}

// ScheduleCancel cancels a scheduled task
func (w *Worker) ScheduleCancel(scheduleId string) (err error) {
	_, err = w.schedules(scheduleId, "cancel").Request("POST", nil)
	return
}
