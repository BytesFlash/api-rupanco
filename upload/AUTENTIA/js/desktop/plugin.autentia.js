/*
//JSON Comando esperado...
//   "paquete":[{"comando":"IniciarSesionLogin","param":"13468309-0"},
//              {"comando":"IniciarSesion","param":"13468309-0"},
//              {"comando":"ParamsInit","param":"Rut,mail"},
//				{"comando":"ParamsSet","param":[{"idx":1,"valor":"0011644656-1"},
//			                                  {"idx":2,"valor":"dupa"}]},
//            	{"comando":"Transaccion", "param":"../DCDA/getmail"			  
//			  	{"comando":"ParamsGet","param":2},
//			  	{"comando":"cerrarSesion"}]}
*/

var PublicCERT = "-----BEGIN CERTIFICATE----- \
MIIDtDCCApwCCQDyBCwa1s/2nzANBgkqhkiG9w0BAQsFADCBmzELMAkGA1UEBhMC \
Q0wxFDASBgNVBAgMC1Byb3ZpZGVuY2lhMREwDwYDVQQHDAhTYW50aWFnbzERMA8G \
A1UECgwIQXV0ZW50aWExEzARBgNVBAsMCklubm92YWNpb24xGDAWBgNVBAMMD011 \
bHRpYnJvd3NlciBKUzEhMB8GCSqGSIb3DQEJARYSc29wb3J0ZUBhY2VwdGEuY29t \
MB4XDTE1MDczMDIxMjYzOVoXDTE2MDcyOTIxMjYzOVowgZsxCzAJBgNVBAYTAkNM \
MRQwEgYDVQQIDAtQcm92aWRlbmNpYTERMA8GA1UEBwwIU2FudGlhZ28xETAPBgNV \
BAoMCEF1dGVudGlhMRMwEQYDVQQLDApJbm5vdmFjaW9uMRgwFgYDVQQDDA9NdWx0 \
aWJyb3dzZXIgSlMxITAfBgkqhkiG9w0BCQEWEnNvcG9ydGVAYWNlcHRhLmNvbTCC \
ASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAK02PVvUOkj5gSANZbsaCsq2 \
5PjSHPTzRbhLsnh4ILJhjsK+22mYSdI9eUD03kxakkYgRXbaaI5Vb/bLLnE92eVB \
ej9D6zOR8H53SXwrIEWtef7mCnQv2Y5l3GQjjk/+yNHs249AQaB3CdHCCuU1FUBo \
pBWm3dLiayBZD/J46KRkwMbA6/qifh2D9PC32fA93AsDC9kHYAoNv0Yeo/0D29PL \
N8M8Qvt1UeI7Zf/l9RhVvxVyIu7pUnYZIykm+TCtLuAdw01UkCUavFDJ96zsIztD \
W0wOpM+lGlWKqMd2LqfWl2iMrB3uHQEV/oFnFM47Ltv3eNabACUcRAInGhCyfXMC \
AwEAATANBgkqhkiG9w0BAQsFAAOCAQEAK2DnMXoPwt+9Ta24oIuUkcSgjUZ/fKiA \
CQ9B/tYS/Di6c5P8w3FW1j5JqWbp9HVxizxqqAKBkEovb9aYT35iSUwKN7ZBxPqC \
VPeNZJGR4tCMhYe4Tdvs007/Ag+YX3DrEbb9e2/DhYUU9eZVfUdyt11mHqspRRAS \
PQ4GK52u8tJ3fNXdm89Vli9J5RvI/Z6W+XfdYyVwSBSTrBjRpJtLcrrXt/nO1jMh \
RPBlKhHD0xmmh3ne+F1GCmcIJXwb8S1TaJer59hR+tpgxOSY6NwKmTPmGoMBLG4J \
oZmeq9HECQEuV3Kg0iq0e1wQAJH7fIZ7y9QXeopudP6eNz/uFwXE5Q== \
-----END CERTIFICATE----- \
";

/*
window.onload = function () {
    if (typeof window.jQuery == 'undefined') {
        var jq = document.createElement('script'); 
            jq.type = 'text/javascript'; 
            jq.src = ('https:' == document.location.protocol ? 'https://' : 'http://') + 'ajax.googleapis.com/ajax/libs/jquery/1/jquery.min.js';
        var sc = document.getElementsByTagName('script')[0];
            sc.parentNode.insertBefore(jq, sc);
    } 

    console.log($('title').html());

}
*/

if (window.console === undefined) {
	window.console = {
		log: function() {},
		error: function() {},
		debug: function() {}
	}
}

if (console.log === undefined) {
	console.log = function() {};
}

if (console.error === undefined) {
	console.error = function() {};
}

if (console.debug === undefined) {
	console.debug = function() {};
}

function _indexOf(array, item) {
	for (var n = 0; n < array.length; n++) {
		if (array[n] === item)
			return n;
	}
	return -1;
}

function _filter(array, _fn) {
	var result = [];
	var item;
	for (item in array) {
		if (_fn(array[item]))
			result.push(array[item]);
	}
	return result;
}

function _trim(s) {
	return s.replace(/^[ \t\r\n]*/, '').replace(/[ \t\r\n]*$/, '');
}

if (! Function.prototype.bind) {
	Function.prototype.bind = function(oThis) {
		if (typeof this !== 'function') {
			// closest thing possible to the ECMAScript 5
			// internal IsCallable function
			throw new TypeError('Function.prototype.bind - what is trying to be bound is not callable');
		}

		var aArgs   = Array.prototype.slice.call(arguments, 1),
			fToBind = this,
			fNOP    = function() {},
			fBound  = function() {
				return fToBind.apply(this instanceof fNOP && oThis
				? this
				: oThis,
				aArgs.concat(Array.prototype.slice.call(arguments)));
			};

		fNOP.prototype = this.prototype;
		fBound.prototype = new fNOP();

		return fBound;
	};
}

function done(){ //on submit function
    console.log('Titulo ventana obtenido desde JQuery '+$('title').html());
}

function load(){ //load jQuery if it isn't already
    window.onload = function(){
        if(window.jQuery === undefined){
            var src = 'https:' == location.protocol ? 'https':'http',
            			script = document.createElement('script');
            script.onload = done;
            script.src = src+'://ajax.googleapis.com/ajax/libs/jquery/1.7.2/jquery.min.js';
            document.getElementsByTagName('body')[0].appendChild(script);

 			script = document.createElement('script');
            script.onload = done;
            script.src = location.href.substring(0, location.href.lastIndexOf("/")+1) + 'js/jquery.blockUI.js';
            document.getElementsByTagName('body')[0].appendChild(script);
			$.blockUI();            
        }else{
            done();
        }
    }
}

if(window.readyState){ //older microsoft browsers
    window.onreadystatechange = function(){
        if(this.readyState == 'complete' || this.readyState == 'loaded'){
            load();
        }
    }
}else{ //modern browsers
    load();
}


function debugLog(msg){
	if (!isIE()) {
		window.console && console.log(msg);	
	}
}

function isIE() {
	var result = false;
	if (navigator.appName.indexOf("Internet Explorer")!=-1) {
		result = true;
	}
	return result;	
}

function getCookie(cname) {
    var name = cname + "=";
    var ca = document.cookie.split(';');
    for(var i=0; i<ca.length; i++) {
        var c = ca[i];
        while (c.charAt(0)==' ') c = c.substring(1);
        if (c.indexOf(name) == 0) return c.substring(name.length,c.length);
    }
    return "";
}

function httpRequest()
{
	var x;
	if (window.XMLHttpRequest)
	  {// code for IE7+, Firefox, Chrome, Opera, Safari
	  	x=new XMLHttpRequest();
	  	debugLog('LOAD:: XMLHttpRequest()');
	  }
	else
	  {// code for IE6, IE5
	  	x=new ActiveXObject("Microsoft.XMLHTTP");
	  	debugLog('LOAD:: ActiveXObject("Microsoft.XMLHTTP")');
	 }	
	 return x;
}


function procesarJSON(data, requerimiento, procesarResultado)
{
	var encodedData = JSON.stringify(data);

	$.ajax ({headers: {
				'Content-type':'text/plain; charset=ISO8859_1'
				// "Content-length": string.length,
				// "Connection":"close"
			},
			type: 'POST',
			url: 'https://plugin.autentia.mb:7777/' + requerimiento,
			success: function(encodedResp){
				var resp = JSON.parse(encodedResp);

				if (exitoAutentia(resp)) {
					/* validamos la firma de la respuesta del multibrowser solo cuando representa un exito de la transaccion */

					//var tmpSignature = JSON.stringify(data).replace(/"signature":"[^"]+"/, '"signature":""');
					var tmpSignature = encodedResp.replace(/"signature":"[^"]+"/, '"signature":""');
					var tmpSignature = tmpSignature.replace(/"token":"[^"]+"/, '"token":""');
					
					var x509 = new X509();
					x509.readCertPEM(PublicCERT);
					var firmaValida = x509.subjectPublicKeyRSA.verifyString(tmpSignature, resp.signature);

					if (! firmaValida) {
						resp = {ParamsGet: {
								ercText: 'Firma de proceso seguro invalido'
							}
						}
					}
				}

				if (errorFatalAutentia(resp)) {
					console.error(resp.ercText);
				}

				procesarResultado(resp);
			},
			error: function(){
				procesarResultado({
					ParamsGet: {
						ercText: 'Error de comunicacion con la componente Autentia.'
					}
				});
			},
			complete: function(){
			},
			dataType: 'text',
			data: encodedData
		});
};

var transaccion={};
var objPaquete=[];
var prmSet=[];
var token='';

function mensajeBloqueo(mensaje, ruedita) {
	var mensajeHtml = '<h3 style="font-family: Arial;">&nbsp;&nbsp;';
	var urigif = location.href.replace(/[^/]+$/, "js/ruedita.gif");
	if (ruedita === true) {
		mensajeHtml += '<img style="vertical-align: middle" src="' + urigif + '" />&nbsp;&nbsp;&nbsp;';
	}
	mensajeHtml += mensaje + '&nbsp;&nbsp;</h3>';

	return {
		message: mensajeHtml,
        centerY: 0,
		css: { 
        	border: 'none', 
        	// padding: '5px', 
        	backgroundColor: '#ffffff', 
        	'-webkit-border-radius': '10px', 
        	'-moz-border-radius': '10px', 
        	opacity: 1, 
        	color: '#000000',
        	top: '10px', 
        	left: '', 
        	right: '10px',
        	width: '400px',
        	"font-family": "Arial",
        	"font-weight": "Bold"
    	}
	}
}

function plgAutentiaJS()
{
	debugLog('Plugin Loaded!!');
};

function excResp(response)
{
	return response;
};

function valueOrDefault(valueFunction, defaultValue) {
	try {
		var value = valueFunction();
		if (value !== undefined)
			return value;
	}
	catch(e) {}
	return defaultValue;
}

function exitoAutentia(respuesta) {
	return valueOrDefault(function() {return respuesta.ParamsGet.erc;}) == "0";
}

function errorFatalAutentia(respuesta) {
	return valueOrDefault(function() {return respuesta.error;}, "") != "";
}

var ProcesoFirmaPDF = function (codDocumento, rutFirmante, nombreFirmante, apellidosFirmante, institucion, notificarFinProceso) {
	this.codDocumento = _filter(codDocumento.split('|'), function (item) { return _trim(item) != ""; });
	this.rutFirmante = rutFirmante;
	this.nombreFirmante = nombreFirmante;
	this.apellidosFirmante = apellidosFirmante;
	this.institucion = institucion;
	this.ctxtId = "";
	this.resultados = [];
	this.notificarFinProceso = notificarFinProceso;
	this.documentoPorFirmar = 0;
	this.lastToken = "";
}

ProcesoFirmaPDF.prototype.getToken = function() {
	this.lastToken = Math.random().toString();
	return this.lastToken;
}

ProcesoFirmaPDF.prototype.solicitarFirmaDocumento = function() {
	var nroDocumentos = this.codDocumento.length;
	if (this.documentoPorFirmar < nroDocumentos) {
		// tenemos documentos por firmar
		// bloqueamos la UI
		$.blockUI(mensajeBloqueo('Procesando documento ' + (this.documentoPorFirmar + 1) + ' de ' + this.codDocumento.length, true));

		var transaccion = {
			paquete: [{
				comando: "pdfXSign",
				doc: this.codDocumento[this.documentoPorFirmar],
				rutFirma: this.rutFirmante + "|" + this.ctxtId,
				nomFirma: this.nombreFirmante,
				apellidosFirma: this.apellidosFirmante,
				inst: this.institucion
			}],
			token: this.getToken()
		}

		procesarJSON(transaccion, 'pdfXSign', this.procesarResultadoFirma.bind(this, new Date().getTime()));
	}
	else {
		// terminamos de firmar todo sin problemas
		$.unblockUI();
		this.notificarFinProceso(true);
	}
}

ProcesoFirmaPDF.prototype.procesarResultadoFirma = function(started, resultado) {
	this.resultados.push(resultado);
	this.documentoPorFirmar += 1;
	var ended = new Date().getTime();
	console.log("pdfXSign: " + (ended - started) + " ms");
	if (exitoAutentia(resultado)) {
		// la firma del documento termino con exito, seguimos firmando
		if (this.ctxtId == "") {
			this.ctxtId = resultado.ParamsGet.ctxtId;
		}
		this.solicitarFirmaDocumento();
	}
	else {
		// hubo un error de firma de un documento,
		// se notifica el error e interrumpe el proceso
		$.unblockUI();
		this.notificarFinProceso(false, resultado);
	}
}

plgAutentiaJS.prototype.pdfXSign = function(codDocumento, rutFirmante, nombreFirmante, ApellidosFirmante, Institucion, token, excResp) {
	var procesoFirma = new ProcesoFirmaPDF(codDocumento, rutFirmante, nombreFirmante, ApellidosFirmante, Institucion, excResp);
	procesoFirma.solicitarFirmaDocumento();
};

plgAutentiaJS.prototype.IniciarSesion = function(rut, token, excResp) {
	objPaquete.push({comando:"IniciarSesion",param:rut});
	var response;
	transaccion.paquete = objPaquete;
	transaccion.token = token;
	/*
	Para su funcionamiento con cross domain con diferentes protocoloes, es decir llamado desde
	https hacia un http en Firefox esta opcion esta bloqueada por defecto, pero se puede habilitar 
	en una opcion que se despliega en la barra de direccion con forma de escudo. En Chrome, solo 
	muestra un warning en la consola del browser.
	*/
	response = procesarJSON(transaccion, 'initSesion', function(response)
	{
		transaccion = {};
		objPaquete = [];
		prmSet = [];
		
		return excResp(response);	
	});		
};

plgAutentiaJS.prototype.IniciarSesionLogin = function(rut, token, excResp) 
{
	objPaquete.push({comando:"IniciarSesionLogin",param:rut});
	var response;
	transaccion.paquete = objPaquete;
	transaccion.token = token;
	/*
	Para su funcionamiento con cross domain con diferentes protocoloes, es decir llamado desde
	https hacia un http en Firefox esta opcion esta bloqueada por defecto, pero se puede habilitar 
	en una opcion que se despliega en la barra de direccion con forma de escudo. En Chrome, solo 
	muestra un warning en la consola del browser.
	*/
	response = procesarJSON(transaccion, 'initSesion', function(response)
	{
		transaccion = {};
		objPaquete = [];
		prmSet = [];
		
		return excResp(response);	
	});	
}

plgAutentiaJS.prototype.Transaccion2 = function(trxName, entrada, salida, hookAutentia, token, excResp)
{
	$.blockUI(mensajeBloqueo('Ejecutando transaccion autentia.'));

	var prmSet = [];
	var ParamInit = [];
    var i = 0;
    var getOffset = 0;

	for (var x in entrada) {
		ParamInit.push(x);
		i += 1;
		prmSet.push({idx: i, valor: entrada[x]});
	}

	for (var x in salida) {
		if (_indexOf(ParamInit, salida[x]) < 0) {
			ParamInit.push(salida[x]);
		}
	}

	objPaquete.push({comando: "ParamsInit", param: ParamInit.join(",")});
	if (prmSet.length) {
		objPaquete.push({comando: "ParamsSet", param: prmSet});
	}					
	objPaquete.push({comando: "Transaccion", param: trxName});

	for(var x in salida) {
		//objPaquete.push({comando: "ParamsGet", param: _indexOf(ParamInit, x) + 1, paramName: x}); // NO FUNCIONA YA QUE 'X' SOLO OBTIENE EL INDICE DEL ARRAY
		objPaquete.push({comando: "ParamsGet", param: _indexOf(ParamInit, salida[x]) + 1, paramName: salida[x]});
	}

	var response, transaccion = {};

	transaccion.paquete = objPaquete;
	transaccion.hookAutentia = hookAutentia;
	transaccion.token = token;
	/*
	Para su funcionamiento con cross domain con diferentes protocoloes, es decir llamado desde
	https hacia un http en Firefox esta opcion esta bloqueada por defecto, pero se puede habilitar 
	en una opcion que se despliega en la barra de direccion con forma de escudo. En Chrome, solo 
	muestra un warning en la consola del browser.
	*/
	response = procesarJSON(transaccion, 'json-handler', function(response)
	{
		transaccion = {};
		objPaquete = [];
		prmSet = [];
		$.unblockUI();
		return excResp(response);	
	});


}

plgAutentiaJS.prototype.CerrarSesion = function(token) 
{
	objPaquete.push({comando:"CerrarSesion", param:""});
	var response;
	transaccion.paquete = objPaquete;
	transaccion.token = token;
	/*
	Para su funcionamiento con cross domain con diferentes protocoloes, es decir llamado desde
	https hacia un http en Firefox esta opcion esta bloqueada por defecto, pero se puede habilitar 
	en una opcion que se despliega en la barra de direccion con forma de escudo. En Chrome, solo 
	muestra un warning en la consola del browser.
	*/
	response = procesarJSON(transaccion, 'closeSesion', function(response)
	{
		transaccion = {};
		objPaquete = [];
		prmSet = [];
		
		return response;	
	});		
}