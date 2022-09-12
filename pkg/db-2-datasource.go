package main

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"time"

	_ "database/sql"

	db2 "github.com/jcnnrts/go_ibm_db"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

// newDatasource returns datasource.ServeOpts.
func newDatasource() datasource.ServeOpts {

	log.DefaultLogger.Warn("Creating new Db2 datasource")

	// creates a instance manager for your plugin. The function passed
	// into `NewInstanceManger` is called when the instance is created
	// for the first time or when a datasource configuration changed.
	im := datasource.NewInstanceManager(newDataSourceInstance)

	ds := &Db2Datasource{
		im: im,
	}

	return datasource.ServeOpts{
		QueryDataHandler:   ds,
		CheckHealthHandler: ds,
	}
}

// Db2Datasource is an example datasource used to scaffold
// new datasource plugins with an backend.
type Db2Datasource struct {
	// The instance manager can help with lifecycle management
	// of datasource instances in plugins. It's not a requirements
	// but a best practice that we recommend that you follow.
	im instancemgmt.InstanceManager
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifer).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (td *Db2Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {

	//Get the instance settings for the current instance of the Db2Datasource.
	instance, err := td.im.Get(req.PluginContext)
	if err != nil {
		log.DefaultLogger.Warn("Failed getting PluginContext")
		return nil, nil
	}

	instSetting, ok := instance.(*instanceSettings)
	if !ok {
		log.DefaultLogger.Warn("Failed getting instance settings")
		return nil, nil
	}

	//Open DB
	instSetting.mu.Lock()
	db := instSetting.pool.Open(instSetting.constr, "SetConnMaxLifetime=0")
	instSetting.mu.Unlock()

	//log.DefaultLogger.Warn("QueryData() - " + instSetting.name)

	response := backend.NewQueryDataResponse()

	// Loop over queries and execute them individually.
	for _, q := range req.Queries {
		// Save the response in a hashmap based on with RefID as identifier
		response.Responses[q.RefID] = td.query(ctx, db, q)
	}

	instSetting.mu.Lock()
	db.Close()
	instSetting.mu.Unlock()

	return response, nil
}

//Query model consists of nothing but a raw query.
type queryModel struct {
	Hide      bool   `json:"hide"`
	QueryText string `json:"queryText"`
}

func (td *Db2Datasource) query(ctx context.Context, db *db2.DBP, query backend.DataQuery) backend.DataResponse {
	//Prepare response objects.
	response := backend.DataResponse{}
	frame := data.NewFrame("response")
	response.Frames = append(response.Frames, frame)

	// Unmarshal the json into our queryModel.
	var qm queryModel
	response.Error = json.Unmarshal(query.JSON, &qm)
	if response.Error != nil {
		return response
	}

	//If query is hidden we don't have to show it.
	if qm.Hide == true {
		return response
	}

	// Run the query
	rows, err := db.Query(qm.QueryText)

	if err != nil {
		log.DefaultLogger.Warn("Query() - Failed running query")
		response.Error = err
		return response
	}

	defer rows.Close()

	//Get names & types of columns, they will be used as names for the series.
	columns, err := rows.ColumnTypes()
	if err != nil {
		log.DefaultLogger.Warn("Query() - Failed to get rows.ColumnTypes()")
		response.Error = err
		return response
	}

	//Make a Field in our frame for every column we have to send back.
	for _, column := range columns {
		frame.Fields = append(frame.Fields, data.NewField(column.Name(), nil, getArrayOfType(column.ScanType())))
	}

	//Make an array of interfaces (values) because we don't know what types the returns are.
	//Then make an array of pointers, because rows.Scan() only takes pointers.
	colPtrs := make([]interface{}, len(columns))
	values := make([]interface{}, len(columns))

	//Point the pointers to our values
	for i := range columns {
		colPtrs[i] = &values[i]
	}

	//Go over each row in the resultset and add its values to the timeseries and the dataseriesMap.
	for rows.Next() {
		err = rows.Scan(colPtrs...)

		if err != nil {
			log.DefaultLogger.Warn("Query() - Failed to do rows.Scan()")
			log.DefaultLogger.Warn(err.Error())

			emptyResponse := backend.DataResponse{}
			emptyResponse.Error = err
			return emptyResponse
		}

		frame.AppendRow(values...)

	}

	return response
}

func getArrayOfType(typ reflect.Type) interface{} {
	switch t := typ.String(); t {
	case "timestamp", "time.Time":
		return []time.Time{}
	case "bigint", "int", "int64":
		return []int64{}
	case "smallint":
		return []int16{}
	case "int32":
		return []int32{}
	case "tinyint":
		return []int8{}
	case "double", "varint", "decimal", "float64":
		return []float64{}
	case "float":
		return []float32{}
	case "string":
		return []*string{}
	default:
		return []*string{}
	}
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (td *Db2Datasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	var status = backend.HealthStatusOk
	var message = "MESSAGE NOT SET YET"

	instance, err := td.im.Get(req.PluginContext)
	if err != nil {
		log.DefaultLogger.Info("Failed getting PluginContext")
		return nil, nil
	}

	instSetting, ok := instance.(*instanceSettings)
	if !ok {
		log.DefaultLogger.Info("Failed getting instance settings")
		return nil, nil
	}

	log.DefaultLogger.Warn("Checkhealth() fired")

	instSetting.mu.Lock()
	db := instSetting.pool.Open(instSetting.constr, "SetConnMaxLifetime=60")
	instSetting.mu.Unlock()
	st, err := db.Prepare("select current timestamp from sysibm.sysdummy1")

	if err != nil {
		log.DefaultLogger.Warn("CheckHealth - Failed on prepare")
		log.DefaultLogger.Warn(err.Error())
	}

	log.DefaultLogger.Warn("CheckHealth - about to run query")
	rows, err := st.Query()

	if err != nil {
		log.DefaultLogger.Warn("CheckHealth - error running query")
		log.DefaultLogger.Warn(err.Error())
	} else {
		if rows != nil {
			log.DefaultLogger.Warn("CheckHealth - getting columns")
			cols, err := rows.Columns()

			if err != nil {
				log.DefaultLogger.Warn("CheckHealth - error getting columns")
				log.DefaultLogger.Warn(err.Error())
			} else {
				log.DefaultLogger.Warn("Number of columns " + cols[0])

				for rows.Next() {
					var tme string

					err := rows.Scan(&tme)
					if err != nil {
						log.DefaultLogger.Warn("CheckHealth - error scanning rows")
						log.DefaultLogger.Warn(err.Error())
					} else {
						log.DefaultLogger.Warn("Current time " + tme)
						message = "Check succesful; current timestamp = " + tme
					}

					rows.Close()
				}
			}
		}
	}

	instSetting.mu.Lock()
	db.Close()
	instSetting.mu.Unlock()

	log.DefaultLogger.Warn("CheckHealth - db closed")

	return &backend.CheckHealthResult{
		Status:  status,
		Message: message,
	}, nil

}

type instanceSettings struct {
	pool   db2.Pool
	mu     sync.Mutex
	constr string
	name   string
}

type myDataSourceOptions struct {
	Host     string
	Port     string
	Database string
	User     string
}

//InstanceFactoryFunc implementation.
func newDataSourceInstance(setting backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	log.DefaultLogger.Warn("newDataSourceInstance()", "data", setting.JSONData)

	// Initialize the Db2 connection pool.
	pl := db2.Pconnect(100)

	// Unload the unsecured JSON data in a myDataSourceOptions struct.
	var dso myDataSourceOptions

	err := json.Unmarshal(setting.JSONData, &dso)
	if err != nil {
		log.DefaultLogger.Warn("error marshaling", "err", err)
		return nil, err
	}

	//Fetch the password from the secured JSON conainer.
	password, _ := setting.DecryptedSecureJSONData["password"]

	constr := fmt.Sprintf("HOSTNAME=%s;PORT=%s;DATABASE=%s;UID=%s;PWD=%s", dso.Host, dso.Port, dso.Database, dso.User, password)

	return &instanceSettings{
		pool:   *pl,
		constr: constr,
		name:   setting.Name,
	}, nil
}

func (s *instanceSettings) Dispose() {
	// Called before creatinga a new instance to allow plugin authors
	// to cleanup.
}
