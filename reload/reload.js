;;;
var tid;
async function checkReload() {
    var uuid = document.body.getAttribute("launchuuid")

    cLaunchUuid = await fetch("/reload/uuid").then( async (response) => {
      return (await response.json())["launchUUID"]
    }).catch( (reason) => document.reload() )

    if(uuid == null) {
        // Get the current UUID
        document.body.setAttribute("launchuuid", cLaunchUuid)
        clearInterval(tid);
        tid = setInterval(checkReload, 5000);
    } else if( uuid != cLaunchUuid ) {
        document.location.reload();
    }
};
(() =>{
        tid = setInterval(checkReload, 2000)
    
})();;;