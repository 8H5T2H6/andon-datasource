package middleware

import (
	"DataSource/config"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bitly/go-simplejson"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func createMongoSession() *mgo.Session {
	session, _ := mgo.Dial(config.MongodbURL)
	// return the connection
	return session
}

func closeMongoSession(Session *mgo.Session) {
	Session.Close()
}

// TestConnection ...
func TestConnection(w http.ResponseWriter, r *http.Request) {
	log.Println("/")
	w.Header().Set("Server", "Grafana Datasource Server")
	w.WriteHeader(200)
	msg := "This is DataSource for iFactory/Andon"
	w.Write([]byte(msg))
}

// GetGroup ...
func GetGroup(w http.ResponseWriter, r *http.Request) {
	log.Println("/group")
	session := createMongoSession()
	db := session.DB(config.MongodbDatabase)
	db.Login(config.MongodbUsername, config.MongodbPassword)
	groupTopoCollection := db.C(config.GroupTopo)
	var results []map[string]interface{}
	var groupNameArray []map[string]string
	groupTopoCollection.Find(bson.M{}).All(&results)
	for _, iV := range results {
		temp := map[string]string{"text": iV["GroupName"].(string), "value": iV["GroupID"].(string)}
		groupNameArray = append(groupNameArray, temp)
	}
	// fmt.Println(groupNameArray)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(groupNameArray)
	closeMongoSession(session)
}

// Search ...
func Search(w http.ResponseWriter, r *http.Request) {
	fmt.Println("/search")
	requestBody, _ := ioutil.ReadAll(r.Body)
	fmt.Println("Body: ", string(requestBody))
	metrics := []string{"EventLatest", "EventList", "EventHist", "LastMonthAbReasonRank", "MTTD", "MTTR", "MTBF", "FactoryMap", "DO-Singlestat", "Event-Singlestat"}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(metrics)
}

// Query ...
func Query(w http.ResponseWriter, r *http.Request) {
	grafnaResponseArray := []map[string]interface{}{}
	log.Println("/query")
	requestBody, _ := simplejson.NewFromReader(r.Body)
	fmt.Println("Body: ", requestBody)
	// fmt.Println(requestBody.Get("targets").MustArray())
	for indexOfTargets := 0; indexOfTargets < len(requestBody.Get("targets").MustArray()); indexOfTargets++ {
		dataType := requestBody.Get("targets").GetIndex(indexOfTargets).Get("type").MustString()
		groupID := requestBody.Get("targets").GetIndex(indexOfTargets).Get("groupID").MustString()
		if strings.HasPrefix(groupID, "$") {
			temp := strings.Split(groupID, "$")
			varName := temp[1]
			groupID = requestBody.Get("scopedVars").Get(varName).Get("value").MustString()
		}
		metrics := requestBody.Get("targets").GetIndex(indexOfTargets).Get("metrics").MustString()
		fromTimeString := requestBody.Get("scopedVars").Get("__from").Get("value").MustString()
		fromUnixTime, _ := strconv.ParseInt(fromTimeString, 10, 64)
		// fmt.Println("FromUnixTime:", fromUnixTime)
		temp := fromUnixTime / 1000
		fromTime := time.Unix(temp, 0)
		toTimeSting := requestBody.Get("scopedVars").Get("__to").Get("value").MustString()
		toUnixTime, _ := strconv.ParseInt(toTimeSting, 10, 64)
		// fmt.Println("ToUnixTime:", toUnixTime)
		temp = toUnixTime / 1000
		toTime := time.Unix(temp, 0)
		fmt.Println("DataType:", dataType, " GroupID:", groupID, " Metrics:", metrics, " From:", fromUnixTime, fromTime.Format(time.RFC3339), " To:", toUnixTime, toTime.Format(time.RFC3339))
		if dataType == "table" {
			switch metrics {
			case "EventLatest":
				grafnaResponseArray = append(grafnaResponseArray, eventLatest(groupID))
			case "EventHist":
				grafnaResponseArray = append(grafnaResponseArray, eventHist(groupID))
			case "EventList":
				grafnaResponseArray = append(grafnaResponseArray, eventList(groupID, fromTime, toTime))
			case "Event-Singlestat":
				grafnaResponseArray = append(grafnaResponseArray, eventSinglestat(groupID))
			case "DO-Singlestat":
				grafnaResponseArray = append(grafnaResponseArray, deviceOverviewSinglestat(groupID))
			case "LastMonthAbReasonRank":
				grafnaResponseArray = append(grafnaResponseArray, lastMonthAbReasonRank(groupID))
			case "FactoryMap":
				grafnaResponseArray = append(grafnaResponseArray, factoryMapTable(groupID))
			}
		} else {
			switch metrics {
			case "MTTD", "MTTR", "MTBF":
				grafnaResponseArray = append(grafnaResponseArray, mtCompute(metrics, groupID))
			case "FactoryMap":
				grafnaResponseArray = factoryMapTimeseries(groupID)
			}
		}
	}
	// jsonStr, _ := json.Marshal(grafnaResponseArray)
	// fmt.Println(string(jsonStr))
	// fmt.Println(grafnaResponseArray)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(grafnaResponseArray)
}

// Table
func eventLatest(groupID string) map[string]interface{} {
	session := createMongoSession()
	db := session.DB(config.MongodbDatabase)
	db.Login(config.MongodbUsername, config.MongodbPassword)
	eventLatestcollection := db.C(config.EventLatest)
	var eventLatestResults []map[string]interface{}
	rows := []interface{}{}
	eventLatestcollection.Pipe([]bson.M{{"$match": bson.M{"GroupID": groupID}}}).All(&eventLatestResults)

	for _, eventLatestResult := range eventLatestResults {

		// for key,value := range eventLatestResult {
		// 	var row []interface{}
		// 	row = append(row, value)
		// }

		var row []interface{}
		row = append(row, eventLatestResult["EventID"])
		row = append(row, eventLatestResult["EventCode"])
		row = append(row, eventLatestResult["Type"])
		row = append(row, eventLatestResult["GroupID"])
		row = append(row, eventLatestResult["GroupName"])
		row = append(row, eventLatestResult["MachineID"])
		row = append(row, eventLatestResult["MachineName"])
		row = append(row, eventLatestResult["AbnormalStartTime"])
		row = append(row, eventLatestResult["ProcessingStatusCode"])
		row = append(row, eventLatestResult["ProcessingProgress"])
		row = append(row, eventLatestResult["AbnormalLastingSecond"])
		row = append(row, eventLatestResult["ShouldRepairTime"])
		row = append(row, eventLatestResult["PlanRepairTime"])
		row = append(row, eventLatestResult["TPCID"])
		row = append(row, eventLatestResult["TPCName"])
		row = append(row, eventLatestResult["PrincipalID"])
		row = append(row, eventLatestResult["PrincipalName"])
		row = append(row, eventLatestResult["AbnormalReason"])
		row = append(row, eventLatestResult["AbnormalSolution"])
		row = append(row, eventLatestResult["AbnormalCode"])
		row = append(row, eventLatestResult["AbnormalPosition"])
		// fmt.Println(row)
		rows = append(rows, row)
	}
	// fmt.Println(rows)

	columns := []map[string]string{
		{"text": "EventID", "type": "string"},
		{"text": "EventCode", "type": "number"},
		{"text": "Type", "type": "string"},
		{"text": "GroupID", "type": "string"},
		{"text": "GroupName", "type": "string"},
		{"text": "MachineID", "type": "string"},
		{"text": "MachineName", "type": "string"},
		{"text": "AbnormalStartTime", "type": "time"},
		{"text": "ProcessingStatusCode", "type": "number"},
		{"text": "ProcessingProgress", "type": "string"},
		{"text": "AbnormalLastingSecond", "type": "number"},
		{"text": "ShouldRepairTime", "type": "time"},
		{"text": "PlanRepairTime", "type": "time"},
		{"text": "TPCID", "type": "string"},
		{"text": "TPCName", "type": "string"},
		{"text": "PrincipalID", "type": "string"},
		{"text": "PrincipalName", "type": "string"},
		{"text": "AbnormalReason", "type": "string"},
		{"text": "AbnormalSolution", "type": "string"},
		{"text": "AbnormalCode", "type": "number"},
		{"text": "AbnormalPosition", "type": "string"},
	}

	grafanaData := map[string]interface{}{
		"columns": columns,
		"rows":    rows,
		"type":    "table",
	}

	closeMongoSession(session)
	return grafanaData
}

func eventList(groupID string, startTime time.Time, endTime time.Time) map[string]interface{} {
	session := createMongoSession()
	db := session.DB(config.MongodbDatabase)
	db.Login(config.MongodbUsername, config.MongodbPassword)
	eventLatestCollection := db.C(config.EventLatest)
	eventHistCollection := db.C(config.EventHist)
	rows := []interface{}{}

	var eventLatestResults []map[string]interface{}
	eventLatestCollection.Pipe([]bson.M{{"$match": bson.M{"GroupID": groupID}}}).All(&eventLatestResults)
	for _, eventLatestResult := range eventLatestResults {

		// for key,value := range eventLatestResult {
		// 	var row []interface{}
		// 	row = append(row, value)
		// }

		var row []interface{}
		row = append(row, eventLatestResult["EventID"])
		row = append(row, eventLatestResult["EventCode"])
		row = append(row, eventLatestResult["Type"])
		row = append(row, eventLatestResult["GroupID"])
		row = append(row, eventLatestResult["GroupName"])
		row = append(row, eventLatestResult["MachineID"])
		row = append(row, eventLatestResult["MachineName"])
		row = append(row, eventLatestResult["AbnormalStartTime"])
		row = append(row, eventLatestResult["ProcessingStatusCode"])
		row = append(row, eventLatestResult["ProcessingProgress"])
		row = append(row, eventLatestResult["AbnormalLastingSecond"])
		row = append(row, eventLatestResult["ShouldRepairTime"])
		row = append(row, eventLatestResult["PlanRepairTime"])
		row = append(row, eventLatestResult["TPCID"])
		row = append(row, eventLatestResult["PrincipalID"])
		row = append(row, eventLatestResult["PrincipalName"])
		row = append(row, eventLatestResult["AbnormalReason"])
		row = append(row, eventLatestResult["AbnormalSolution"])
		row = append(row, eventLatestResult["AbnormalCode"])
		row = append(row, eventLatestResult["AbnormalPosition"])
		// fmt.Println(row)
		rows = append(rows, row)
	}

	var eventHistResults []map[string]interface{}
	eventHistCollection.Pipe([]bson.M{{"$match": bson.M{"GroupID": groupID}}, {"$match": bson.M{"AbnormalStartTime": bson.M{"$gte": startTime, "$lt": endTime}}}}).All(&eventHistResults)
	for _, eventHistResult := range eventHistResults {

		// for key,value := range eventLatestResult {
		// 	var row []interface{}
		// 	row = append(row, value)
		// }

		var row []interface{}
		row = append(row, eventHistResult["EventID"])
		row = append(row, eventHistResult["EventCode"])
		row = append(row, eventHistResult["Type"])
		row = append(row, eventHistResult["GroupID"])
		row = append(row, eventHistResult["GroupName"])
		row = append(row, eventHistResult["MachineID"])
		row = append(row, eventHistResult["MachineName"])
		row = append(row, eventHistResult["AbnormalStartTime"])
		row = append(row, eventHistResult["ProcessingStatusCode"])
		row = append(row, eventHistResult["ProcessingProgress"])
		row = append(row, eventHistResult["AbnormalLastingSecond"])
		row = append(row, eventHistResult["ShouldRepairTime"])
		row = append(row, eventHistResult["PlanRepairTime"])
		row = append(row, eventHistResult["TPCID"])
		row = append(row, eventHistResult["PrincipalID"])
		row = append(row, eventHistResult["PrincipalName"])
		row = append(row, eventHistResult["AbnormalReason"])
		row = append(row, eventHistResult["AbnormalSolution"])
		row = append(row, eventHistResult["AbnormalCode"])
		row = append(row, eventHistResult["AbnormalPosition"])
		// fmt.Println(row)
		rows = append(rows, row)
	}

	// fmt.Println(rows)

	columns := []map[string]string{
		{"text": "EventID", "type": "string"},
		{"text": "EventCode", "type": "number"},
		{"text": "Type", "type": "string"},
		{"text": "GroupID", "type": "string"},
		{"text": "GroupName", "type": "string"},
		{"text": "MachineID", "type": "string"},
		{"text": "MachineName", "type": "string"},
		{"text": "AbnormalStartTime", "type": "time"},
		{"text": "ProcessingStatusCode", "type": "number"},
		{"text": "ProcessingProgress", "type": "string"},
		{"text": "AbnormalLastingSecond", "type": "number"},
		{"text": "ShouldRepairTime", "type": "time"},
		{"text": "PlanRepairTime", "type": "time"},
		{"text": "TPCID", "type": "string"},
		{"text": "PrincipalID", "type": "string"},
		{"text": "PrincipalName", "type": "string"},
		{"text": "AbnormalReason", "type": "string"},
		{"text": "AbnormalSolution", "type": "string"},
		{"text": "AbnormalCode", "type": "number"},
		{"text": "AbnormalPosition", "type": "string"},
	}

	grafanaData := map[string]interface{}{
		"columns": columns,
		"rows":    rows,
		"type":    "table",
	}

	closeMongoSession(session)
	return grafanaData
}

func eventSinglestat(groupID string) map[string]interface{} {
	session := createMongoSession()
	db := session.DB(config.MongodbDatabase)
	db.Login(config.MongodbUsername, config.MongodbPassword)
	eventLatestCollection := db.C(config.EventLatest)
	eventHistCollection := db.C(config.EventHist)
	tpcListCollection := db.C(config.TPCList)
	nowTime := time.Now().In(config.TaipeiTimeZone)
	starttimeFS := fmt.Sprintf("%02d-%02d-%02dT00:00:00+08:00", nowTime.Year(), nowTime.Month(), nowTime.Day())
	starttimeT, _ := time.Parse(time.RFC3339, starttimeFS)
	endtimeFS := fmt.Sprintf("%02d-%02d-%02dT23:59:59+08:00", nowTime.Year(), nowTime.Month(), nowTime.Day())
	endtimeT, _ := time.Parse(time.RFC3339, endtimeFS)
	var rows []interface{}

	var eventHistResults []map[string]interface{}
	eventHistCollection.Pipe([]bson.M{{"$match": bson.M{"GroupID": groupID}}, {"$match": bson.M{"AbnormalStartTime": bson.M{"$gte": starttimeT, "$lte": endtimeT}}}, {"$match": bson.M{"$expr": bson.M{"$or": []bson.M{{"ProcessingStatusCode": 5}, {"ProcessingStatusCode": 6}}}}}}).All(&eventHistResults)
	// fmt.Println("EventHistResults:", eventHistResults)
	numOfCompleted := len(eventHistResults)

	var tpcListResults []map[string]interface{}
	tpcListCollection.Find(bson.M{"GroupID": groupID}).All(&tpcListResults)
	// fmt.Println("TPCListResults:", tpcListResults)
	numOfTPC := len(tpcListResults)

	var eventLatestResults []map[string]interface{}
	eventLatestCollection.Find(bson.M{"GroupID": groupID}).All(&eventLatestResults)

	numOfLatestEvents := len(eventLatestResults)
	// fmt.Println("NumOfEvents -> ", numOfEvents)
	numOfTotalEvents := len(eventLatestResults) + len(eventHistResults)

	numOfCompletedToday := 0
	numOfOverdue := 0
	numOfWaiting := 0
	numOfInProgressing := 0
	for _, eventLatestResult := range eventLatestResults {
		if eventLatestResult["PlanRepairTime"] != nil {
			planRepairTime := eventLatestResult["PlanRepairTime"].(time.Time)
			if planRepairTime.After(starttimeT) && planRepairTime.Before(endtimeT) {
				numOfCompletedToday++
			}
		}
		processingStatusCode := eventLatestResult["ProcessingStatusCode"]
		if processingStatusCode == 0 || processingStatusCode == 1 || processingStatusCode == 4 {
			numOfWaiting++
			if processingStatusCode == 4 {
				numOfOverdue++
			}
		} else if processingStatusCode == 3 {
			numOfInProgressing++
		}
	}
	// eventLatestCollection.Pipe([]bson.M{{"$match": bson.M{"GroupID": groupID}}, {"$match": bson.M{"PlanRepairTime": bson.M{"$gte": starttimeT, "$lte": endtimeT}}}}).All(&eventLatestResults)
	// numOfCompletedToday := len(eventLatestResults)
	// fmt.Println("NumOfCompletedToday -> ", numOfCompletedToday)

	// eventLatestCollection.Pipe([]bson.M{{"$match": bson.M{"GroupID": groupID}}, {"$match": bson.M{"ProcessingStatusCode": 4}}}).All(&eventLatestResults)
	// numOfOverdue := len(eventLatestResults)
	// fmt.Println("NumOfOverdue -> ", numOfOverdue)

	var row []interface{}
	row = append(row, numOfTPC)
	row = append(row, numOfTotalEvents)
	row = append(row, numOfLatestEvents)
	row = append(row, numOfCompletedToday)
	row = append(row, numOfOverdue)
	row = append(row, numOfWaiting)
	row = append(row, numOfInProgressing)
	row = append(row, numOfCompleted)
	rows = append(rows, row)
	// fmt.Println(rows)
	columns := []map[string]string{
		{"text": "NumOfTPC", "type": "number"},
		{"text": "NumOfTotalEvents", "type": "number"},
		{"text": "NumOfLatestEvents", "type": "number"},
		{"text": "CompletedToday", "type": "number"},
		{"text": "Overdue", "type": "number"},
		{"text": "Waiting", "type": "number"},
		{"text": "InProgressing", "type": "number"},
		{"text": "Completed", "type": "number"},
	}
	grafanaData := map[string]interface{}{
		"columns": columns,
		"rows":    rows,
		"type":    "table",
	}
	closeMongoSession(session)
	return grafanaData
}

func deviceOverviewSinglestat(groupID string) map[string]interface{} {
	session := createMongoSession()
	db := session.DB(config.MongodbDatabase)
	db.Login(config.MongodbUsername, config.MongodbPassword)
	statisticCollection := db.C(config.Statistic)
	var result map[string]interface{}
	var rows []interface{}
	statisticCollection.Pipe([]bson.M{{"$match": bson.M{"GroupID": groupID}}, {"$match": bson.M{"Dashboard": "DO"}}}).One(&result)
	// fmt.Println("Result: ", result)

	var row []interface{}
	row = append(row, result["Total"])
	row = append(row, result["Run"])
	row = append(row, result["Idle"])
	row = append(row, result["Down"])
	row = append(row, result["Off"])
	row = append(row, result["SumFailureCost"])
	rows = append(rows, row)
	// fmt.Println(rows)
	columns := []map[string]string{
		{"text": "Total", "type": "number"},
		{"text": "Run", "type": "number"},
		{"text": "Idle", "type": "number"},
		{"text": "Down", "type": "number"},
		{"text": "Off", "type": "number"},
		{"text": "SumFailureCost", "type": "number"},
	}
	grafanaData := map[string]interface{}{
		"columns": columns,
		"rows":    rows,
		"type":    "table",
	}
	closeMongoSession(session)
	return grafanaData
}

func eventHist(groupID string) map[string]interface{} {
	session := createMongoSession()
	db := session.DB(config.MongodbDatabase)
	db.Login(config.MongodbUsername, config.MongodbPassword)
	eventHistCollection := db.C(config.EventHist)
	var results []map[string]interface{}
	rows := []interface{}{}
	eventHistCollection.Pipe([]bson.M{{"$match": bson.M{"GroupID": groupID}}, {"$match": bson.M{"EventCode": 1}}}).All(&results)
	for indexOfResult := 0; indexOfResult < len(results); indexOfResult++ {
		var result map[string]interface{}
		result = results[indexOfResult]
		// fmt.Println("Result: ", result)
		var row []interface{}
		row = append(row, result["EventID"])
		row = append(row, result["EventCode"])
		row = append(row, result["Type"])
		row = append(row, result["GroupID"])
		row = append(row, result["GroupName"])
		row = append(row, result["MachineID"])
		row = append(row, result["MachineName"])
		row = append(row, result["AbnormalStartTime"])
		row = append(row, result["ProcessingStatusCode"])
		row = append(row, result["ProcessingProgress"])
		row = append(row, result["AbnormalLastingSecond"])
		row = append(row, result["ShouldRepairTime"])
		row = append(row, result["PlanRepairTime"])
		row = append(row, result["PrincipalID"])
		row = append(row, result["PrincipalName"])
		row = append(row, result["AbnormalReason"])
		row = append(row, result["AbnormalSolution"])
		row = append(row, result["AbnormalCode"])
		row = append(row, result["AbnormalPosition"])
		// fmt.Println(row)
		rows = append(rows, row)
	}
	// fmt.Println(rows)
	columns := []map[string]string{
		{"text": "EventID", "type": "string"},
		{"text": "EventCode", "type": "number"},
		{"text": "Type", "type": "string"},
		{"text": "GroupID", "type": "string"},
		{"text": "GroupName", "type": "string"},
		{"text": "MachineID", "type": "string"},
		{"text": "MachineName", "type": "string"},
		{"text": "AbnormalStartTime", "type": "time"},
		{"text": "ProcessingStatusCode", "type": "number"},
		{"text": "ProcessingProgress", "type": "string"},
		{"text": "AbnormalLastingSecond", "type": "number"},
		{"text": "ShouldRepairTime", "type": "time"},
		{"text": "PlanRepairTime", "type": "time"},
		{"text": "PrincipalID", "type": "string"},
		{"text": "PrincipalName", "type": "string"},
		{"text": "AbnormalReason", "type": "string"},
		{"text": "AbnormalSolution", "type": "string"},
		{"text": "AbnormalCode", "type": "number"},
		{"text": "AbnormalPosition", "type": "string"},
	}
	grafanaData := map[string]interface{}{
		"columns": columns,
		"rows":    rows,
		"type":    "table",
	}
	closeMongoSession(session)
	return grafanaData
}

func lastMonthAbReasonRank(groupID string) map[string]interface{} {
	rows := []interface{}{}
	// var starttimeT, endtimeT time.Time
	// nowTime := time.Now().In(config.TaipeiTimeZone)
	session := createMongoSession()
	db := session.DB(config.MongodbDatabase)
	db.Login(config.MongodbUsername, config.MongodbPassword)
	eventHistCollection := db.C(config.EventHist)
	var results []map[string]interface{}
	// year := nowTime.Year()
	// month := int(nowTime.Month())
	// endtimeFS := fmt.Sprintf("%02d-%02d-01T00:00:00+08:00", year, month)
	// endtimeT, _ = time.Parse(time.RFC3339, endtimeFS)
	// if month-1 == 0 {
	// 	month = 12
	// 	year = year - 1
	// } else {
	// 	month = month - 1
	// }
	// starttimeFS := fmt.Sprintf("%02d-%02d-01T00:00:00+08:00", year, month)
	// starttimeT, _ = time.Parse(time.RFC3339, starttimeFS)
	eventHistCollection.Pipe([]bson.M{{"$match": bson.M{"GroupID": groupID}}, {"$group": bson.M{"_id": "$AbnormalReason", "count": bson.M{"$sum": 1}}}, {"$sort": bson.M{"count": -1}}}).All(&results)
	// fmt.Println(results)
	recordCount := 0
	for _, result := range results {
		if result["_id"] != nil && recordCount < 10 {
			// fmt.Println(result)
			row := []interface{}{result["_id"], result["count"]}
			rows = append(rows, row)
			recordCount++
		}
	}
	// fmt.Println(rows)
	columns := []map[string]string{
		{"text": "AbnormalReason", "type": "string"},
		{"text": "Count", "type": "float"},
	}
	grafanaData := map[string]interface{}{
		"columns": columns,
		"rows":    rows,
		"type":    "table",
	}
	// fmt.Println(grafanaData)
	return grafanaData
}

func factoryMapTable(groupID string) map[string]interface{} {
	grafanaData := map[string]interface{}{}
	session := createMongoSession()
	db := session.DB(config.MongodbDatabase)
	db.Login(config.MongodbUsername, config.MongodbPassword)
	machineRawDataCollection := db.C(config.MachineRawData)
	var results []map[string]interface{}
	machineRawDataCollection.Pipe([]bson.M{{"$match": bson.M{"GroupID": groupID}}}).All(&results)
	columns := []map[string]string{}
	rows := []interface{}{}
	row := []interface{}{}
	for indexOfResult := 0; indexOfResult < len(results); indexOfResult++ {
		var result map[string]interface{}
		result = results[indexOfResult]
		// fmt.Println("Result: ", result)
		row = append(row, result["StatusLay1Value"])
		// fmt.Println(row)
		columns = append(columns, map[string]string{"text": result["MachineName"].(string), "type": "number"})
	}
	rows = append(rows, row)
	fmt.Println(columns, rows)
	grafanaData = map[string]interface{}{
		"columns": columns,
		"rows":    rows,
		"type":    "table",
	}
	closeMongoSession(session)
	return grafanaData
}

// Timeseries
func mtCompute(computeType string, groupID string) map[string]interface{} {
	var starttimeT, endtimeT time.Time
	nowTime := time.Now().In(config.TaipeiTimeZone)
	var results []map[string]interface{}
	var datapoints []interface{}
	session := createMongoSession()
	db := session.DB(config.MongodbDatabase)
	db.Login(config.MongodbUsername, config.MongodbPassword)
	eventHistCollection := db.C(config.EventHist)
	year := nowTime.Year()
	month := int(nowTime.Month())
	for monthIndex := 0; monthIndex < 12; monthIndex++ {
		var datapoint []interface{}
		if monthIndex == 0 {
			endtimeT = nowTime
		} else {
			endtimeT = starttimeT
		}
		starttimeFS := fmt.Sprintf("%02d-%02d-01T00:00:00+08:00", year, month)
		starttimeT, _ = time.Parse(time.RFC3339, starttimeFS)
		// fmt.Println("StartTime:", starttimeT, "EndTime:", endtimeT)
		var sumOfSecond float64
		if computeType == "MTTD" {
			eventHistCollection.Pipe([]bson.M{{"$match": bson.M{"GroupID": groupID}}, {"$match": bson.M{"AbnormalStartTime": bson.M{"$gte": starttimeT, "$lt": endtimeT}}}, {"$match": bson.M{"RepairStartTime": bson.M{"$gte": starttimeT, "$lt": endtimeT}}}}).All(&results)
			for _, result := range results {
				// fmt.Println(result)
				ast := result["AbnormalStartTime"].(time.Time)
				rst := result["RepairStartTime"].(time.Time)
				// fmt.Println("   ", "AST:", ast, " RST:", rst)
				mttd := rst.Sub(ast).Seconds()
				// fmt.Println("       ", "MTTD:", mttd)
				sumOfSecond = sumOfSecond + mttd
			}
		} else if computeType == "MTTR" {
			eventHistCollection.Pipe([]bson.M{{"$match": bson.M{"GroupID": groupID}}, {"$match": bson.M{"RepairStartTime": bson.M{"$gte": starttimeT, "$lt": endtimeT}}}, {"$match": bson.M{"CompleteTime": bson.M{"$gte": starttimeT, "$lt": endtimeT}}}}).All(&results)
			// fmt.Println("Results:", results)
			for _, result := range results {
				rst := result["RepairStartTime"].(time.Time)
				rct := result["CompleteTime"].(time.Time)
				// fmt.Println(" RST:", rst, " RCT:", rct)
				mttr := rct.Sub(rst).Seconds()
				// fmt.Println("MTTR:", mttr)
				sumOfSecond = sumOfSecond + mttr
			}
		} else if computeType == "MTBF" {
			eventHistCollection.Pipe([]bson.M{{"$match": bson.M{"GroupID": groupID}}, {"$match": bson.M{"RepairStartTime": bson.M{"$gte": starttimeT, "$lt": endtimeT}}}, {"$match": bson.M{"CompleteTime": bson.M{"$gte": starttimeT, "$lt": endtimeT}}}}).All(&results)
			var sumOfFailureTime float64
			for _, result := range results {
				ast := result["AbnormalStartTime"].(time.Time)
				rct := result["CompleteTime"].(time.Time)
				// fmt.Println("AST:", ast, " RCT:", rct)
				failureTime := rct.Sub(ast).Seconds()
				// fmt.Println("FailureTime:", failureTime)
				sumOfFailureTime = sumOfFailureTime + failureTime
			}
			// fmt.Println("SumOfFailureTime:", sumOfFailureTime)
			totalTime := 864000.0 // 30 Day * 8 Hour * 60 Minute * 60 Second
			sumOfSecond = totalTime - sumOfFailureTime
			// fmt.Println("SumOfSecond:", sumOfSecond)
		}

		if len(results) == 0 {
			datapoint = []interface{}{0.0, starttimeT.Unix() * 1000}

		} else {
			datapoint = []interface{}{(sumOfSecond / float64(len(results))), starttimeT.Unix() * 1000}
		}
		datapoints = append(datapoints, datapoint)

		if (month - 1) == 0 {
			month = 12
			year = year - 1
		} else {
			month = month - 1
		}
	}
	for i, j := 0, len(datapoints)-1; i < j; i, j = i+1, j-1 {
		datapoints[i], datapoints[j] = datapoints[j], datapoints[i]
	}

	grafanaData := map[string]interface{}{
		"target":     computeType,
		"datapoints": datapoints,
	}
	// fmt.Println(grafanaData)
	return grafanaData
}

func factoryMapTimeseries(groupID string) []map[string]interface{} {
	grafnaResponseArray := []map[string]interface{}{}
	session := createMongoSession()
	db := session.DB(config.MongodbDatabase)
	db.Login(config.MongodbUsername, config.MongodbPassword)
	machineRawDataCollection := db.C(config.MachineRawData)
	tpcListCollection := db.C(config.TPCList)
	eventLatestCollection := db.C(config.EventLatest)

	var tpcListResults []map[string]interface{}
	tpcListCollection.Pipe([]bson.M{{"$match": bson.M{"GroupID": groupID}}}).All(&tpcListResults)

	for _, tpcListResult := range tpcListResults {
		var datapoints []interface{}
		var datapoint []interface{}

		var eventLatestResults []map[string]interface{}
		eventLatestCollection.Pipe([]bson.M{{"$match": bson.M{"TPCID": tpcListResult["TPCID"]}}, {"$match": bson.M{"EventCode": 2}}}).All(&eventLatestResults)

		datapoint = []interface{}{len(eventLatestResults), time.Now().In(config.TaipeiTimeZone).Unix() * 1000}
		datapoints = append(datapoints, datapoint)
		grafanaData := map[string]interface{}{
			"target":     tpcListResult["TPCName"],
			"datapoints": datapoints,
		}
		grafnaResponseArray = append(grafnaResponseArray, grafanaData)
	}

	var machineRawDataResults []map[string]interface{}
	machineRawDataCollection.Pipe([]bson.M{{"$match": bson.M{"GroupID": groupID}}}).All(&machineRawDataResults)

	for indexOfMachineRawDataResult := 0; indexOfMachineRawDataResult < len(machineRawDataResults); indexOfMachineRawDataResult++ {
		var datapoints []interface{}
		var datapoint []interface{}
		var result map[string]interface{}
		result = machineRawDataResults[indexOfMachineRawDataResult]
		// fmt.Println("Result: ", result)
		if result["ManualEvent"].(int) > 0 {
			datapoint = []interface{}{2000, result["Timestamp"].(time.Time).Unix() * 1000}
		} else {
			datapoint = []interface{}{result["StatusLay1Value"], result["Timestamp"].(time.Time).Unix() * 1000}
		}
		// fmt.Println(row)
		datapoints = append(datapoints, datapoint)
		grafanaData := map[string]interface{}{
			"target":     result["MachineName"],
			"datapoints": datapoints,
		}
		grafnaResponseArray = append(grafnaResponseArray, grafanaData)
	}
	closeMongoSession(session)
	return grafnaResponseArray
}

//Test ...
// func Test(w http.ResponseWriter, r *http.Request) {
