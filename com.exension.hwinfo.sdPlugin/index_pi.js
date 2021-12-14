// this is our global websocket, used to communicate from/to Stream Deck software
// and some info about our plugin, as sent by Stream Deck software
var websocket = null,
  uuid = null,
  actionInfo = {},
  inInfo = {},
  runningApps = [],
  isQT = navigator.appVersion.includes("QtWebEngine"),
  onchangeevt = "onchange"; // 'oninput'; // change this, if you want interactive elements act on any change, or while they're modified

function connectElgatoStreamDeckSocket(inPort, inUUID, inRegisterEvent, inInfo, inActionInfo) {
  uuid = inUUID;
  // please note: the incoming arguments are of type STRING, so
  // in case of the inActionInfo, we must parse it into JSON first
  actionInfo = JSON.parse(inActionInfo); // cache the info
  inInfo = JSON.parse(inInfo);
  websocket = new WebSocket("ws://localhost:" + inPort);

  /** Since the PI doesn't have access to native settings
   * Stream Deck sends some color settings to PI
   * We use these to adjust some styles (e.g. highlight-colors for checkboxes)
   */
  addDynamicStyles(inInfo.colors, "connectSocket");
  initPropertyInspector(5);

  // if connection was established, the websocket sends
  // an 'onopen' event, where we need to register our PI
  websocket.onopen = function () {
    var json = {
      event: inRegisterEvent,
      uuid: inUUID,
    };
    // register property inspector to Stream Deck
    websocket.send(JSON.stringify(json));
    sendValueToPlugin("propertyInspectorConnected", "property_inspector");
  };

  websocket.onmessage = function (evt) {
    // Received message from Stream Deck
    var jsonObj = JSON.parse(evt.data);
    var event = jsonObj["event"];
    if (
      "boolean" === typeof getPropFromString(jsonObj, "payload.error") &&
      event === "sendToPropertyInspector"
    ) {
      if (jsonObj.payload.error === true) {
        document.querySelector("#ui").style = "display:none";
        document.querySelector("#error").style = "display:block";
      } else if (jsonObj.payload.message === "show_ui") {
        document.querySelector("#ui").style = "display:block";
        document.querySelector("#error").style = "display:none";
        sendValueToPlugin("propertyInspectorConnected", "property_inspector");
      }
    }
    if (
      getPropFromString(jsonObj, "payload.sensors") &&
      event === "sendToPropertyInspector"
    ) {
      addSensors(
        document.querySelector("#sensorSelect"),
        jsonObj.payload.sensors,
        jsonObj.payload.settings
      );
    }
    if (
      getPropFromString(jsonObj, "payload.readings") &&
      event === "sendToPropertyInspector"
    ) {
      addReadings(
        document.querySelector("#readingSelect"),
        jsonObj.payload.readings,
        jsonObj.payload.settings
      );
    }
    if (getPropFromString(jsonObj, "payload.settings")) {
      var settings = jsonObj.payload.settings;
      if (settings.min === 0 && settings.max === 0) {
        // don't show 0, 0 min/max
      } else {
        document.querySelector("#min").value = settings.min;
        document.querySelector("#max").value = settings.max;
      }
      document.querySelector("#format input").value = settings.format;
      document.querySelector("#divisor input").value = settings.divisor || "";
      if (
        settings.format.length > 0 ||
        (settings.divisor && settings.divisor.length > 0)
      ) {
        var attr = document.createAttribute("open");
        attr.value = "open";
        document
          .querySelector("#advanced_details")
          .attributes.setNamedItem(attr);
      }
      if (settings.foregroundColor !== "") {
        document.querySelector("#foreground").value = settings.foregroundColor;
      }
      if (settings.backgroundColor !== "") {
        document.querySelector("#background").value = settings.backgroundColor;
      }
      if (settings.highlightColor !== "") {
        document.querySelector("#highlight").value = settings.highlightColor;
      }
      if (settings.valueTextColor !== "") {
        document.querySelector("#valuetext").value = settings.valueTextColor;
      }
      if (settings.titleFontSize !== "") {
        document.querySelector("#titleFontSize input").value =
          settings.titleFontSize || 10.5;
      }
      if (settings.valueFontSize !== "") {
        document.querySelector("#valueFontSize input").value =
          settings.valueFontSize || 10.5;
      }
    }
  };
}

function sortBy(key) {
  return function (a, b) {
    if (a[key] > b[key]) return 1;
    if (b[key] > a[key]) return -1;
    return 0;
  };
}

function addSensors(el, sensors, settings) {
  var i;
  for (i = el.options.length - 1; i >= 0; i--) {
    el.remove(i);
  }

  el.removeAttribute("disabled");

  var option = document.createElement("option");
  option.text = "Choose a sensor";
  option.disabled = true;
  if (settings.isValid !== true) {
    option.selected = true;
  }
  el.add(option);
  var sortByName = sortBy("name");
  sensors.sort(sortByName).forEach((s) => {
    var option = document.createElement("option");
    option.text = s.name;
    option.value = s.uid;
    if (settings.isValid === true && settings.sensorUid === s.uid) {
      option.selected = true;
      setTimeout(function () {
        var event = new Event("change");
        el.dispatchEvent(event);
      }, 0);
    }
    el.add(option);
  });
}

function addReadings(el, readings, settings) {
  var i;
  for (i = el.options.length - 1; i >= 0; i--) {
    el.remove(i);
  }

  el.removeAttribute("disabled");

  var option = document.createElement("option");
  option.text = "Choose a reading";
  option.disabled = true;
  if (settings.isValid !== true) {
    option.selected = true;
  }
  el.add(option);

  var sortByLabel = sortBy("label");
  var maxL = 0;
  readings.sort(sortByLabel).forEach((r) => {
    var l = r.prefix.length;
    if (l > maxL) {
      maxL = l;
    }
  });
  readings.sort(sortByLabel).forEach((r) => {
    var option = document.createElement("option");
    option.style = "white-space: pre";
    var spaces = "&nbsp;";
    for (i = 0; i < maxL - r.prefix.length; ++i) {
      spaces += "&nbsp;";
    }
    option.innerHTML = `${r.prefix}${spaces}${r.label}`;
    option.value = r.id;
    if (settings.isValid === true && settings.readingId === r.id) {
      option.selected = true;
    }
    el.add(option);
  });
}

function initPropertyInspector(initDelay) {
  prepareDOMElements(document);
}

function revealSdpiWrapper() {
  const el = document.querySelector(".sdpi-wrapper");
  el && el.classList.remove("hidden");
}

// openUrl in default browser
function openUrl(url) {
  if (websocket && websocket.readyState === 1) {
    const json = {
      event: "openUrl",
      payload: {
        url: url,
      },
    };
    websocket.send(JSON.stringify(json));
  }
}

// our method to pass values to the plugin
function sendValueToPlugin(value, param) {
  if (websocket && websocket.readyState === 1) {
    const json = {
      action: actionInfo["action"],
      event: "sendToPlugin",
      context: uuid,
      payload: {
        [param]: value,
      },
    };
    websocket.send(JSON.stringify(json));
  }
}

if (!isQT) {
  document.addEventListener("DOMContentLoaded", function () {
    initPropertyInspector(100);
  });
}

/** the beforeunload event is fired, right before the PI will remove all nodes */
window.addEventListener("beforeunload", function (e) {
  e.preventDefault();
  sendValueToPlugin("propertyInspectorWillDisappear", "property_inspector");
  // Don't set a returnValue to the event, otherwise Chromium with throw an error.  // e.returnValue = '';
});

/** the pagehide event is fired, when the view disappears */
/*
window.addEventListener('pagehide', function (event) {
    console.log('%c%s','background: green; font-size: 22px; font-weight: bold;','window --->> pagehide.');
    sendValueToPlugin('propertyInspectorPagehide', 'property_inspector');

});
*/

/** the unload event is fired, when the PI will finally disappear */
/*
window.addEventListener('unload', function (event) {
    console.log('%c%s','background: orange; font-size: 22px; font-weight: bold;','window --->> onunload.');
    sendValueToPlugin('propertyInspectorDisconnected', 'property_inspector');
});
*/

/** if you prefer, you can apply these listeners to PI's body, like so:
 *
 * <body onpagehide="sendValueToPlugin('propertyInspectorPagehide', 'property_inspector');">
 *
 * <body onunload="sendValueToPlugin('propertyInspectorDisconnected', 'property_inspector');">
 */

/** CREATE INTERACTIVE HTML-DOM
 * where elements can be clicked or act on their 'change' event.
 * Messages are then processed using the 'handleSdpiItemClick' method below.
 */

function prepareDOMElements(baseElement) {
  baseElement = baseElement || document;
  Array.from(baseElement.querySelectorAll(".sdpi-item-value")).forEach(
    (el, i) => {
      const elementsToClick = [
        "BUTTON",
        "OL",
        "UL",
        "TABLE",
        "METER",
        "PROGRESS",
        "CANVAS",
      ].includes(el.tagName);
      const evt = elementsToClick ? "onclick" : onchangeevt || "onchange";
      // console.log(el.type, el.tagName, elementsToClick, el, evt);

      /** Look for <input><span> combinations, where we consider the span as label for the input
       * we don't use `labels` for that, because a range could have 2 labels.
       */
      const inputGroup = el.querySelectorAll("input, span");
      if (inputGroup.length === 2) {
        const offs = inputGroup[0].tagName === "INPUT" ? 1 : 0;
        inputGroup[offs].innerText = inputGroup[1 - offs].value;
        inputGroup[1 - offs]["oninput"] = function () {
          inputGroup[offs].innerText = inputGroup[1 - offs].value;
        };
      }
      /** We look for elements which have an 'clickable' attribute
       * we use these e.g. on an 'inputGroup' (<span><input type="range"><span>) to adjust the value of
       * the corresponding range-control
       */
      Array.from(el.querySelectorAll(".clickable")).forEach((subel, subi) => {
        subel["onclick"] = function (e) {
          handleSdpiItemClick(e.target, subi);
        };
      });
      el[evt] = function (e) {
        handleSdpiItemClick(e.target, i);
      };
    }
  );

  baseElement.querySelectorAll("textarea").forEach((e) => {
    const maxl = e.getAttribute("maxlength");
    e.targets = baseElement.querySelectorAll(`[for='${e.id}']`);
    if (e.targets.length) {
      let fn = () => {
        for (let x of e.targets) {
          x.innerText = maxl
            ? `${e.value.length}/${maxl}`
            : `${e.value.length}`;
        }
      };
      fn();
      e.onkeyup = fn;
    }
  });
}

function handleSdpiItemClick(e, idx) {
  /** Following items are containers, so we won't handle clicks on them */
  if (["OL", "UL", "TABLE"].includes(e.tagName)) {
    return;
  }
  // console.log('--- handleSdpiItemClick ---', e, `type: ${e.type}`, e.tagName, `inner: ${e.innerText}`);

  /** SPANS are used inside a control as 'labels'
   * If a SPAN element calls this function, it has a class of 'clickable' set and is thereby handled as
   * clickable label.
   */

  if (e.tagName === "SPAN") {
    const inp = e.parentNode.querySelector("input");
    if (e.getAttribute("value")) {
      return inp && (inp.value = e.getAttribute("value"));
    }
  }

  const selectedElements = [];
  const isList = ["LI", "OL", "UL", "DL", "TD"].includes(e.tagName);
  const sdpiItem = e.closest(".sdpi-item");
  const sdpiItemGroup = e.closest(".sdpi-item-group");
  let sdpiItemChildren = isList
    ? sdpiItem.querySelectorAll(e.tagName === "LI" ? "li" : "td")
    : sdpiItem.querySelectorAll(".sdpi-item-child > input");

  if (isList) {
    const siv = e.closest(".sdpi-item-value");
    if (!siv.classList.contains("multi-select")) {
      for (let x of sdpiItemChildren) x.classList.remove("selected");
    }
    if (!siv.classList.contains("no-select")) {
      e.classList.toggle("selected");
    }
  }

  if (sdpiItemGroup && !sdpiItemChildren.length) {
    for (let x of ["input", "meter", "progress"]) {
      sdpiItemChildren = sdpiItemGroup.querySelectorAll(x);
      if (sdpiItemChildren.length) break;
    }
  }

  if (e.selectedIndex) {
    idx = e.selectedIndex;
  } else {
    sdpiItemChildren.forEach((ec, i) => {
      if (ec.classList.contains("selected")) {
        selectedElements.push(ec.innerText);
      }
      if (ec === e) idx = i;
    });
  }

  const returnValue = {
    key: e.id || sdpiItem.id,
    value: isList
      ? e.innerText
      : e.value
        ? e.type === "file"
          ? decodeURIComponent(e.value.replace(/^C:\\fakepath\\/, ""))
          : e.value
        : e.getAttribute("value"),
    group: sdpiItemGroup ? sdpiItemGroup.id : false,
    index: idx,
    selection: selectedElements,
    checked: e.checked,
  };

  /** Just simulate the original file-selector:
   * If there's an element of class '.sdpi-file-info'
   * show the filename there
   */
  if (e.type === "file") {
    const info = sdpiItem.querySelector(".sdpi-file-info");
    if (info) {
      const s = returnValue.value.split("/").pop();
      info.innerText =
        s.length > 28
          ? s.substr(0, 10) + "..." + s.substr(s.length - 10, s.length)
          : s;
    }
  }

  sendValueToPlugin(returnValue, "sdpi_collection");
}

function updateKeyForDemoCanvas(cnv) {
  sendValueToPlugin(
    {
      key: "your_canvas",
      value: cnv.toDataURL(),
    },
    "sdpi_collection"
  );
}

/** Stream Deck software passes system-highlight color information
 * to Property Inspector. Here we 'inject' the CSS styles into the DOM
 * when we receive this information. */

function addDynamicStyles(clrs, fromWhere) {
  const node =
    document.getElementById("#sdpi-dynamic-styles") ||
    document.createElement("style");
  if (!clrs.mouseDownColor)
    clrs.mouseDownColor = fadeColor(clrs.highlightColor, -100);
  const clr = clrs.highlightColor.slice(0, 7);
  const clr1 = fadeColor(clr, 100);
  const clr2 = fadeColor(clr, 60);
  const metersActiveColor = fadeColor(clr, -60);

  node.setAttribute("id", "sdpi-dynamic-styles");
  node.innerHTML = `

    input[type="radio"]:checked + label span,
    input[type="checkbox"]:checked + label span {
        background-color: ${clrs.highlightColor};
    }

    input[type="radio"]:active:checked + label span,
    input[type="radio"]:active + label span,
    input[type="checkbox"]:active:checked + label span,
    input[type="checkbox"]:active + label span {
      background-color: ${clrs.mouseDownColor};
    }

    input[type="radio"]:active + label span,
    input[type="checkbox"]:active + label span {
      background-color: ${clrs.buttonPressedBorderColor};
    }

    td.selected,
    td.selected:hover,
    li.selected:hover,
    li.selected {
      color: white;
      background-color: ${clrs.highlightColor};
    }

    .sdpi-file-label > label:active,
    .sdpi-file-label.file:active,
    label.sdpi-file-label:active,
    label.sdpi-file-info:active,
    input[type="file"]::-webkit-file-upload-button:active,
    button:active {
      background-color: ${clrs.buttonPressedBackgroundColor};
      color: ${clrs.buttonPressedTextColor};
      border-color: ${clrs.buttonPressedBorderColor};
    }

    ::-webkit-progress-value,
    meter::-webkit-meter-optimum-value {
        background: linear-gradient(${clr2}, ${clr1} 20%, ${clr} 45%, ${clr} 55%, ${clr2})
    }

    ::-webkit-progress-value:active,
    meter::-webkit-meter-optimum-value:active {
        background: linear-gradient(${clr}, ${clr2} 20%, ${metersActiveColor} 45%, ${metersActiveColor} 55%, ${clr})
    }
    `;
  document.body.appendChild(node);
}

/** UTILITIES */

/** Helper function to construct a list of running apps
 * from a template string.
 * -> information about running apps is received from the plugin
 */

function sdpiCreateList(el, obj, cb) {
  if (el) {
    el.style.display = obj.value.length ? "block" : "none";
    Array.from(document.querySelectorAll(`.${el.id}`)).forEach((subel, i) => {
      subel.style.display = obj.value.length ? "flex" : "none";
    });
    if (obj.value.length) {
      el.innerHTML = `<div class="sdpi-item" ${obj.type ? `class="${obj.type}"` : ""
        } id="${obj.id || window.btoa(new Date().getTime().toString()).substr(0, 8)
        }">
            <div class="sdpi-item-label">${obj.label || ""}</div>
            <ul class="sdpi-item-value ${obj.selectionType ? obj.selectionType : ""
        }">
                    ${obj.value.map((e) => `<li>${e.name}</li>`).join("")}
                </ul>
            </div>`;
      setTimeout(function () {
        prepareDOMElements(el);
        if (cb) cb();
      }, 10);
      return;
    }
  }
  if (cb) cb();
}

/** get a JSON property from a (dot-separated) string
 * Works on nested JSON, e.g.:
 * jsn = {
 *      propA: 1,
 *      propB: 2,
 *      propC: {
 *          subA: 3,
 *          subB: {
 *             testA: 5,
 *             testB: 'Hello'
 *          }
 *      }
 *  }
 *  getPropFromString(jsn,'propC.subB.testB') will return 'Hello';
 */
const getPropFromString = (jsn, str, sep = ".") => {
  const arr = str.split(sep);
  return arr.reduce(
    (obj, key) => (obj && obj.hasOwnProperty(key) ? obj[key] : undefined),
    jsn
  );
};

/*
    Quick utility to lighten or darken a color (doesn't take color-drifting, etc. into account)
    Usage:
    fadeColor('#061261', 100); // will lighten the color
    fadeColor('#200867'), -100); // will darken the color
*/
function fadeColor(col, amt) {
  const min = Math.min,
    max = Math.max;
  const num = parseInt(col.replace(/#/g, ""), 16);
  const r = min(255, max((num >> 16) + amt, 0));
  const g = min(255, max((num & 0x0000ff) + amt, 0));
  const b = min(255, max(((num >> 8) & 0x00ff) + amt, 0));
  return "#" + (g | (b << 8) | (r << 16)).toString(16).padStart(6, 0);
}
