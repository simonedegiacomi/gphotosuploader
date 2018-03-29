function find_WIZ_global_data(elm) {
  if (elm.nodeType == Node.ELEMENT_NODE || elm.nodeType == Node.DOCUMENT_NODE) {
    for (var i=0; i < elm.childNodes.length; i++) {
      // recursively call self
      find_WIZ_global_data(elm.childNodes[i]);
    }
  }

  if (elm.nodeType == Node.TEXT_NODE) {
    if (elm.nodeValue.startsWith("window.WIZ_global_data")) {
      var jsonString = elm.nodeValue.replace("window.WIZ_global_data = ", "");
      jsonString = jsonString.slice(0, -1);
      var wiz = JSON.parse(jsonString);
      chrome.runtime.sendMessage({id: wiz["S06Grb"]});
    }
  }
}

find_WIZ_global_data(document);
