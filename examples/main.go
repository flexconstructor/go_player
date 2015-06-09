package main
import(
	"log"
	"github.com/flexconstructor/GoPlayer"
	"text/template"
	"net/http"

)

var homeTempl = template.Must(template.ParseFiles("./src/github.com/flexconstructor/GoPlayer/examples/index.html"))
var playerTemplate=template.Must(template.ParseFiles("./src/github.com/flexconstructor/GoPlayer/examples/player.html"))
var jsTemplate=template.Must(template.ParseFiles("./src/github.com/flexconstructor/GoPlayer/examples/js/megapix-image.js"))
var preloaderTemplate= template.Must(template.ParseFiles("./src/github.com/flexconstructor/GoPlayer/examples/js/heartcode-canvasloader-min.js"))
var rtmp_url string="rtmp://localhost:1935/tv"


func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", 404)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	log.Print("get page")
	homeTempl.Execute(w, r.Host)
}

func getStreamHandler(w http.ResponseWriter, r *http.Request){
	if r.URL.Path != "/get-stream" {
		http.Error(w, "Not found", 404)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	streamname:=r.FormValue("streamname");
	if(streamname != "") {
		go GoPlayer.NewGoPlayer().Run(streamname,NewGPLogger())
	}else{
		http.Error(w, "Stream Not found", 404)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	playerTemplate.Execute(w,r.Host);
	//homeTempl.Execute(w, r.Host)

}

func jsHandler(w http.ResponseWriter, r *http.Request){

	jsTemplate.Execute(w,r.Host);
}

func preloaderHandler(w http.ResponseWriter, r *http.Request){
	preloaderTemplate.Execute(w,r.Host);
}


func main() {
	log.Print("Hello Go Player")

	http.HandleFunc("/", serveHome)
	http.HandleFunc("/get-stream",getStreamHandler)
	http.HandleFunc("/js/megapix-image.js",jsHandler)
	http.HandleFunc("/js/heartcode-canvasloader-min.js",preloaderHandler)
	err := http.ListenAndServe(":9095", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}



}
