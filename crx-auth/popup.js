var auth = {
  "cookies": [],
  "persistantParameters": {
    "userId": ""
  }
};

chrome.tabs.query({
  active: true,
  currentWindow: true
}, function(tabs) {
  var tab = tabs[0];

  chrome.cookies.getAll({}, function (cookies) {

    var cookieNames = ["OTZ", "CONSENT", "SID", "APISID", "SAPISID", "HSID", "NID", "SSID"];
    var cookieDomains = [".google.com", "photos.google.com"];

    auth["cookies"].length = 0;
    for (var i in cookies) {

      var cookie = cookies[i];
      if (cookieNames.indexOf(cookie.name) == -1) {
        continue;
      }
      if (cookieDomains.indexOf(cookie.domain) == -1) {
        continue;
      }

      var cookieAuth = {};
      cookieAuth["Name"] = cookie.name;
      cookieAuth["Value"] = cookie.value;
      cookieAuth["Domain"] = cookie.domain;
      cookieAuth["HttpOnly"] = cookie.httpOnly;
      cookieAuth["Secure"] = cookie.secure;
      cookieAuth["Path"] = cookie.path;

      auth["cookies"].push(cookieAuth);

    }
    chrome.tabs.executeScript(null, {file: "getid.js"});

  });
});

chrome.runtime.onMessage.addListener(function(request, sender) {
  auth["persistantParameters"]["userId"] = request.id;
  document.write("<pre>");
  document.write(JSON.stringify(auth, null, 2));
  document.write("</pre>");
});
