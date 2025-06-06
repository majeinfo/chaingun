package manager

import (
	"bufio"
	"embed"
	"encoding/json"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/majeinfo/chaingun/reporter"
	"github.com/majeinfo/chaingun/viewer"
	log "github.com/sirupsen/logrus"
)

const (
	fname_suffix = ".data"
)

var (
	repositoryDir string

	// Script names played by the Players
	scriptNames = make(map[string]string)
	scriptName  string
	injectors   []string

	webUrl string = "/web/"

	//go:embed web/*
	content embed.FS
)

func redirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, webUrl, 301)
}

// Start creates the HTTP server and creates the Web Interface
func Start(mgrAddr string, reposdir string, prelaunched_injectors string) error {
	if len(prelaunched_injectors) > 0 {
		injectors = strings.Split(prelaunched_injectors, ",")
	} else {
		injectors = make([]string, 0)
	}
	repositoryDir = reposdir

	mux := http.NewServeMux()
	mux.HandleFunc("/get_injectors", getInjectors)
	mux.HandleFunc("/clean_results", cleanResults)
	mux.HandleFunc("/store_results/", storeResults)
	mux.HandleFunc("/merge_results/", mergeResults)
	mux.HandleFunc("/show_results/", showResults)
	mux.HandleFunc("/rebuild_graphs/", rebuildGraphs)
	mux.Handle("/results/", http.FileServer(http.Dir(".")))
	sub_fs, _ := fs.Sub(content, "web")
	mux.Handle("/", http.FileServer(http.FS(sub_fs)))
	//mux.HandleFunc("/", redirect)

	log.Fatal(http.ListenAndServe(mgrAddr, mux))
	return nil
}

func getInjectors(w http.ResponseWriter, r *http.Request) {
	log.Debugf("getInjectors called")

	jData, err := json.Marshal(injectors)
	if err != nil {
		log.Fatalf("Json Marshaling failed with %v (%s)", injectors, err)
	}
	log.Debugf(string(jData))
	w.Header().Set("Content-Type", "application/json")
	w.Write(jData)
}

func cleanResults(w http.ResponseWriter, r *http.Request) {
	log.Debugf("cleanResults called")
	scriptNames = make(map[string]string)
	scriptName = ""
	sendJSONResponse(w, "OK", "", "")
}

func storeResults(w http.ResponseWriter, r *http.Request) {
	log.Debugf("storeResults called, urlpath=%s", r.URL.Path)

	// The Request Path looks like :
	// /store_results/<repository>/<resultname>/<scriptname>
	parts := strings.Split(r.URL.Path[1:], "/")
	log.Debugf("%v %d", parts, len(parts))
	if len(parts) != 4 {
		sendJSONErrorResponse(w, "Error", "Malformed URL Path")
		return
	}

	// Creates the repository directory if needed
	targetDir := repositoryDir + "/results/" + parts[1] // <repository>
	stat, err := os.Stat(targetDir)
	log.Debugf("Repository Directory is '%s'", targetDir)

	if os.IsNotExist(err) {
		log.Debugf("Must create the Repository Directory")
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			sendJSONErrorResponse(w, "Error", err.Error())
			return
		}
	} else if stat.Mode().IsRegular() {
		sendJSONErrorResponse(w, "Error", "Datadir already exists as a file!")
		return
	}

	// Write the results
	fname := targetDir + "/" + parts[2] + fname_suffix // <resultname>
	log.Debugf("Write results in file %s", fname)
	file, err := os.Create(fname)
	defer file.Close()
	if err != nil {
		sendJSONErrorResponse(w, "Error", err.Error())
		return
	}
	responseBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		sendJSONErrorResponse(w, "Error", err.Error())
		return
	}

	if _, err := file.Write(responseBody); err != nil {
		sendJSONErrorResponse(w, "Error", err.Error())
		return
	}

	scriptNames[parts[2]] = parts[3] // scriptname
	scriptName = parts[3]

	sendJSONResponse(w, "OK", "", "")
}

func mergeResults(w http.ResponseWriter, r *http.Request) {
	log.Debugf("mergeResults called, urlpath=%s", r.URL.Path)

	// The Request Path looks like :
	// /merge_results/<repository>
	parts := strings.Split(r.URL.Path[1:], "/")
	if len(parts) != 2 {
		sendJSONErrorResponse(w, "Error", "Malformed URL Path")
		return
	}

	// Creates the merged file
	repo_dir := repositoryDir + "/results/" + parts[1] // <repository>
	if err := _mergeResults(repo_dir); err != nil {
		sendJSONErrorResponse(w, "Error", err.Error())
		return
	}

	merged_name := "merged.csv"
	fname := repo_dir + "/" + merged_name

	// Create metadata files
	scriptnames := []string{scriptName}
	if err := reporter.WriteMetadata(time.Now(), time.Now(), repo_dir, scriptnames); err != nil {
		log.Error(err.Error())
	}

	// Build graphs...
	if err := viewer.BuildGraphs(fname, parts[1], repo_dir); err != nil {
		log.Errorf("BuildGraphs failed: %s", err)
		sendJSONErrorResponse(w, "Error", err.Error())
	} else {
		sendJSONResponse(w, "OK", "", repo_dir+"/index.html") // TODO: link_url missing
	}
}

func _mergeResults(repo_dir string) error {
	log.Debugf("_mergeResults called")

	// Creates the merged file
	merged_name := "merged.csv"
	fname := repo_dir + "/" + merged_name

	log.Debugf("Creates merged file '%s'", fname)
	mergedFile, err := os.Create(fname)
	//defer mergedFile.Close() // No, because the BuildGraphs function will use it before the end of the function !
	if err != nil {
		mergedFile.Close()
		return err
	}

	// Loop on repo_dir content
	files, err := ioutil.ReadDir(repo_dir)
	if err != nil {
		mergedFile.Close()
		return err
	}

	first_file := true
	for _, file := range files {
		log.Debugf("Found file %s", file.Name())
		//if file.Name() == "merged.csv" || file.Name() == "." || file.Name() == ".." {
		if strings.LastIndex(file.Name(), fname_suffix) == -1 {
			continue
		}
		first_line := true

		log.Debugf("Open file '%s'", repo_dir+"/"+file.Name())
		file, err := os.Open(repo_dir + "/" + file.Name())
		if err != nil {
			mergedFile.Close()
			return err
		}
		//defer file.Close()	// No: too late

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			if first_line && !first_file {
				first_line = false
				continue
			}

			if _, err := mergedFile.WriteString(scanner.Text() + "\n"); err != nil {
				file.Close()
				mergedFile.Close()
				return err
			}
		}
		first_file = false
		file.Close()

		if err := scanner.Err(); err != nil {
			mergedFile.Close()
			return err
		}
	}
	mergedFile.Close()

	return nil
}

func rebuildGraphs(w http.ResponseWriter, r *http.Request) {
	log.Debugf("rebuildGraphs called, urlpath=%s", r.URL.Path)

	// The Request Path looks like :
	// /rebuild_graphs/<repository>
	parts := strings.Split(r.URL.Path[1:], "/")
	if len(parts) != 2 {
		sendJSONErrorResponse(w, "Error", "Malformed URL Path")
		return
	}

	// many .data files or single data file ?
	pathfile := repositoryDir + "/results/" + parts[1] + "/*.data"
	matches, _ := filepath.Glob(pathfile)
	if matches == nil {
		log.Debugf("Single build in %s", pathfile)
		//err := viewer.BuildGraphs(outputFile, scriptName, outputDir)
	} else {
		log.Debugf("Multiple build in %s", pathfile)
		mergeResults(w, r)
	}
}

func showResults(w http.ResponseWriter, r *http.Request) {
	log.Debugf("showResults called: %s", r.URL.Path)
	doc1 := heredoc.Doc(`
	<!DOCTYPE html>
	<html>
	<head>
	
	<meta charset="UTF-8" />
	<meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no"/>
	<meta http-equiv="X-UA-Compatible" content="IE=edge">
	
	<title>Chaingun Management Interface</title>
	
	<link href="/static/css/bootstrap.css" rel="stylesheet" type="text/css">
	<link href="/static/vendor/metisMenu/metisMenu.css" rel="stylesheet" type="text/css">
	<link href="/static/css/sb-admin-2.css" rel="stylesheet" type="text/css">
	<link href="/static/vendor/morrisjs/morris.css" rel="stylesheet" type="text/css">
	<link href="/static/css/jquery.dataTables.min.css" rel="stylesheet" type="text/css">
	<link href="/static/css/dataTables.checkboxes.css" rel="stylesheet" type="text/css">
	<link href="/static/vendor/font-awesome/css/font-awesome.css" rel="stylesheet" type="text/css">
	
	</head>
	<body>
	
	<div id="wrapper">
		<div id="page-wrapper" style="margin-left: 0px;">
			<div class="row">
				<div class="col-lg-12">
				</div>
			</div>
	
			<!-- /.row -->
			<div class="row">
				<div class="col-lg-12">
					<div class="panel panel-default">
						<div class="panel-heading">
							<div class="xcontainer">
								<div class="row">
									<div class="col-lg-4">
										<span class="lead text-left">Results Management</span>
									</div>
									<div class="col-lg-8 text-right">
									</div>
								</div>
							</div>
						</div>
						<!-- /.panel-heading -->
						<div class="panel-body">
                        <table width="100%" class="table table-striped table-bordered table-hover" id="results-table">
                            <thead>
                                <tr>
                                    <th class="text-center">Directory Name</th>
                                    <th class="center">Actions</th>
                                </tr>
                            </thead>
                            <tbody>	
	`)

	doc2 := heredoc.Doc(`
                            </tbody>
                        </table>
                        <!-- /.table-responsive -->
                    </div>
                    <!-- /.panel-body -->
                </div>
                <!-- /.panel -->
            </div>
            <!-- /.col-lg-12 -->
        </div>
    </div>
    <!-- /#page-wrapper -->
</div>  <!-- wrapper -->

<!-- Modal -->
<div id="modalMsg" class="modal fade" role="dialog">
    <div class="modal-dialog">

        <!-- Modal content-->
        <div class="modal-content">
            <div class="modal-header alert-info">
                <button type="button" class="close" data-dismiss="modal">&times;</button>
                <h4 class="modal-title">Message</h4>
            </div>
            <div class="modal-body" id="modal-body"></div>
            <div class="modal-footer">
                <button type="button" class="btn btn-primary" data-dismiss="modal">Close</button>
            </div>
        </div>

    </div>
</div>

<!-- jQuery -->
<script src="/static/js/jquery.min.js"></script>
<script src="/static/js/highcharts.js"></script>
<script src="/static/js/exporting.js"></script>
<script src="/static/js/jquery.dataTables.min.js"></script>
<script src="/static/js/dataTables.checkboxes.min.js"></script>
<script src="/static/js/popper.min.js"></script>
<script src="/static/js/bootstrap.min.js"></script>
<script src="/static/vendor/metisMenu/metisMenu.js"></script>
<script src="/static/vendor/raphael/raphael.js"></script>
<script src="/static/vendor/morrisjs/morris.js"></script>
<script src="/static/dist/js/sb-admin-2.js"></script>

<script>
function viewResults(result_name) {
	window.open('/results/' + result_name, '_blank');
}

function rebuildGraphs(result_name) {
	/* call merge_result if merged.csv or many .data files
	   then buildgraph */
	$.getJSON('/rebuild_graphs/' + result_name, function( data ) {
		var mbody = $("#modal-body");
		mbody.html("The Graphs have been rebuilt !");
		$("#modalMsg").modal();	
	});
}
</script>

</body>
</html>
	`)

	w.Write([]byte(doc1))

	// Loop on content of /results directory
	targetDir := repositoryDir + "/results/"
	files, err := ioutil.ReadDir(targetDir)
	if err != nil {
		w.Write([]byte("Error"))
		w.Write([]byte(err.Error()))
		return
	}

	for _, file := range files {
		stat, err := os.Stat(targetDir + "/" + file.Name())
		if err != nil || !stat.IsDir() {
			continue
		}
		line := heredoc.Docf(`
								<tr class="even gradeA">
                                    <td>%s</td>
                                    <td class="center">
                                        <button type="button" class="btn btn-primary" onclick="viewResults('%s')">View</button>
                                        <button type="button" class="btn btn-primary" onclick="rebuildGraphs('%s')">Rebuild Graphs</button>
                                    </td>
								</tr>
								`, file.Name(), file.Name(), file.Name())
		w.Write([]byte(line))
	}

	w.Write([]byte(doc2))
}

func sendJSONErrorResponse(w http.ResponseWriter, status string, msg string) {
	log.Debugf("sendJSONErrorResponse called")
	sendJSONResponse(w, status, msg, "")
}

func sendJSONResponse(w http.ResponseWriter, status string, msg string, link_url string) {
	log.Debugf("sendJSONResponse called")

	data := map[string]string{
		"status":   status,
		"msg":      msg,
		"link_url": link_url,
	}
	jData, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("Json Marshaling failed with %v (%s)", data, err)
	}
	log.Debugf(string(jData))
	w.Header().Set("Content-Type", "application/json")
	w.Write(jData)
}
