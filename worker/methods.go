package worker

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"time"
)

type Schedule struct {
	CodeName string         `json:"code_name"`
	Name     string         `json:"name"`
	Payload  string         `json:"payload"`
	Delay    *time.Duration `json:"delay"`
	Priority *int           `json:"priority"`
	RunEvery *int           `json:"run_every"`
	RunTimes *int           `json:"run_times"`
	StartAt  *time.Time     `json:"start_at"`
	EndAt    *time.Time     `json:"end_at"`
}

type ScheduleInfo struct {
	CodeName    string    `json:"code_name"`
	Id          string    `json:"id"`
	Msg         string    `json:"msg"`
	ProjectId   string    `json:"project_id"`
	Status      string    `json:"status"`
	RunCount    int       `json:"run_count"`
	RunTimes    int       `json:"run_times"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	NextStart   time.Time `json:"next_start"`
	LastRunTime time.Time `json:"last_run_time"`
	StartAt     time.Time `json:"start_at"`
	EndAt       time.Time `json:"end_at"`
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
	Duration      int       `json:"duration"`
	RunTimes      int       `json:"run_times"`
	Timeout       int       `json:"timeout"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
}

type CodeSource map[string][]byte // map[pathInZip]code

type Code struct {
	Name     string
	Runtime  string
	FileName string
	Source   CodeSource
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
	page = clamp(page, 0, 100)
	perPage = clamp(perPage, 1, 100)
	// TODO: find a nice way to use the url package for that
	err = w.getJSON(fmt.Sprintf("codes?page=%d&perPage=%d", page, perPage), &out)
	if err != nil {
		return
	}
	return out["codes"], nil
}

// CodePackageUpload uploads a code package
func (w *Worker) CodePackageUpload(code Code) (id string, err error) {
	client := http.Client{}
	// TODO: find a nice way to use the url package for that
	uri := fmt.Sprintf("%s%d/projects/%s/%s", w.BaseURL, w.ApiVersion, w.ProjectId, "codes")

	body := &bytes.Buffer{}
	mWriter := multipart.NewWriter(body)

	// write meta-data
	mMetaWriter, err := mWriter.CreateFormField("data")
	if err != nil {
		return
	}
	jEncoder := json.NewEncoder(mMetaWriter)
	err = jEncoder.Encode(map[string]string{
		"name":      code.Name,
		"runtime":   code.Runtime,
		"file_name": code.FileName,
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

	req, err := http.NewRequest("POST", uri, body)
	if err != nil {
		return
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Encoding", "gzip/deflate")
	req.Header.Set("Authorization", "OAuth "+w.Token)
	req.Header.Set("Content-Type", mWriter.FormDataContentType())
	req.Header.Set("User-Agent", w.UserAgent)

	// dumpRequest(req) NOTE: never do this here, it breaks stuff
	response, err := client.Do(req)
	if err != nil {
		return
	}
	if response.StatusCode != httpOk {
		return "", resToErr(response)
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
	err = w.getJSON("codes/"+codeId, &out)
	if err != nil {
		return
	}
	return out, nil
}

// CodePackageDelete deletes a code package
func (w *Worker) CodePackageDelete(codeId string) (err error) {
	_, err = w.request("DELETE", "codes/"+codeId, nil)
	return err
}

// CodePackageDownload downloads a code package
func (w *Worker) CodePackageDownload(codeId string) (code Code, err error) {
	out := Code{}
	err = w.getJSON("codes/"+codeId+"/download", &out)
	if err != nil {
		return
	}
	return out, nil
}

// CodePackageRevisions lists the revisions of a code pacakge
func (w *Worker) CodePackageRevisions(codeId string) (code Code, err error) {
	out := Code{}
	err = w.getJSON("codes/"+codeId+"/revisions", &out)
	if err != nil {
		return
	}
	return out, nil
}

func (w *Worker) TaskList() (tasks []TaskInfo, err error) {
	out := map[string][]TaskInfo{}
	err = w.getJSON("tasks", &out)
	if err != nil {
		return
	}
	return out["tasks"], nil
}

// TaskQueue queues a task
func (w *Worker) TaskQueue(tasks ...Task) (taskIds []string, err error) {
	allTasks := make([]map[string]interface{}, 0, len(tasks))

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

		allTasks = append(allTasks, thisTask)
	}

	res, err := w.post("tasks", allTasks)
	if err != nil {
		return
	}

	data := struct {
		Tasks []struct {
			Id string `json:"id"`
		} `json:"tasks"`
		Msg string `json:"msg"`
	}{}

	err = json.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		return
	}

	for _, task := range data.Tasks {
		taskIds = append(taskIds, task.Id)
	}

	return
}

// TaskInfo gives info about a given task
func (w *Worker) TaskInfo(taskId string) (task TaskInfo, err error) {
	out := TaskInfo{}
	err = w.getJSON("tasks/"+taskId, &out)
	if err != nil {
		return
	}
	return out, nil
}

func (w *Worker) TaskLog(taskId string) (log []byte, err error) {
	res, err := w.request("GET", "tasks/"+taskId+"/log", nil)
	if err != nil {
		return
	}

	log, err = ioutil.ReadAll(res.Body)
	return
}

// TaskCancel cancels a Task
func (w *Worker) TaskCancel(taskId string) (err error) {
	_, err = w.request("POST", "tasks/"+taskId+"/cancel", nil)
	return
}

// TaskProgress sets a Task's Progress
func (w *Worker) TaskProgress(taskId string, progress int) (err error) { return }

// TaskQueueWebhook queues a Task from a Webhook
func (w *Worker) TaskQueueWebhook() (err error) { return }

// ScheduleList lists Scheduled Tasks
func (w *Worker) ScheduleList() (schedules []ScheduleInfo, err error) {
	out := map[string][]ScheduleInfo{}
	err = w.getJSON("schedules", &out)
	if err != nil {
		return
	}
	return out["schedules"], nil
}

// Schedule a Task
func (w *Worker) Schedule(schedules ...Schedule) (scheduleIds []string, err error) {
	allSchedules := make([]map[string]interface{}, 0, len(schedules))

	for _, schedule := range schedules {
		sm := map[string]interface{}{
			"code_name": schedule.CodeName,
			"name":      schedule.Name,
			"payload":   schedule.Payload,
		}
		if schedule.Delay != nil {
			sm["delay"] = (*schedule.Delay).Seconds()
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
		if schedule.EndAt != nil {
			sm["end_at"] = *schedule.EndAt
		}
		allSchedules = append(allSchedules, sm)
	}

	res, err := w.post("schedules", allSchedules)

	data := struct {
		Schedules []struct {
			Id string `json:"id"`
		} `json:"schedules"`
		Msg string `json:"msg"`
	}{}

	err = json.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		return
	}

	scheduleIds = make([]string, 0, len(data.Schedules))

	for _, schedule := range data.Schedules {
		scheduleIds = append(scheduleIds, schedule.Id)
	}

	return
}

// ScheduleInfo gets info about a scheduled task
func (w *Worker) ScheduleInfo(scheduleId string) (info ScheduleInfo, err error) {
	info = ScheduleInfo{}
	err = w.getJSON("schedules/"+scheduleId, &info)
	if err != nil {
		return
	}
	return info, nil
}

// ScheduleCancel cancels a scheduled task
func (w *Worker) ScheduleCancel(scheduleId string) (err error) {
	_, err = w.request("POST", "schedules/"+scheduleId+"/cancel", nil)
	return
}
