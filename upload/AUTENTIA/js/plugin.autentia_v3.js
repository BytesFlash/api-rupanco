/*
//JSON Comando esperado...
//   "paquete":[{"comando": "IniciarSesionLogin", "param": "13468309-0"},
//              {"comando": "IniciarSesion", "param": "13468309-0"},
//              {"comando": "ParamsInit", "param":"Rut,mail"},
//              {"comando": "ParamsSet", "param": [{"idx": 1, "valor": "0011644656-1"},
//                                                 {"idx": 2, "valor": "texto"}]},
//              {"comando": "Transaccion", "param": "../DCDA/getmail"
//              {"comando": "ParamsGet", "param": 2},
//              {"comando": "cerrarSesion"}]}
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

/**
 * BEGIN: method binding tools
 */
/**
 * bind function substitute
 */
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

/**
 * END: method binding tools
 */

/**
 * dummy console for browser that don't have one
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

/**
 * some supplementary funcions for browser without them
 */
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

function done(){
    //on submit function
    //console.log('Titulo ventana obtenido desde JQuery '+$('title').html());
}

function load(){ //load jQuery if it isn't already
    window.onload = function() {
        if (window.jQuery === undefined) {
            var protocol = 'https:' == location.protocol ? 'https': 'http';
            var script = document.createElement('script');
            script.onload = done;
            script.src = protocol + '://ajax.googleapis.com/ajax/libs/jquery/1.7.2/jquery.min.js';
            document.getElementsByTagName('body')[0].appendChild(script);

            script = document.createElement('script');
            script.onload = done;
            //script.src = location.href.substring(0, location.href.lastIndexOf("/")+1) + 'js/jquery.blockUI.js';
            document.getElementsByTagName('body')[0].appendChild(script);
            JQuery.blockUI();
        } else {
            done();
        }
    }
}

if (window.readyState) { //older microsoft browsers
    window.onreadystatechange = function(){
        if(this.readyState == 'complete' || this.readyState == 'loaded'){
            load();
        }
    }
} else { //modern browsers
    load();
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

function httpRequest() {
    var x;
    if (window.XMLHttpRequest) {// code for IE7+, Firefox, Chrome, Opera, Safari
        x=new XMLHttpRequest();
        console.debug('LOAD:: XMLHttpRequest()');
    } else {// code for IE6, IE5
        x=new ActiveXObject("Microsoft.XMLHTTP");
        console.debug('LOAD:: ActiveXObject("Microsoft.XMLHTTP")');
    }
    return x;
}

function FirmaInvalidaMB(resp, encodedResp) {
    var LtmpSignature = encodedResp.replace(/"signature":"[^"]+"/, '"signature":""');
    var Lx509 = new X509();
    Lx509.readCertPEM(PublicCERT);
    var LfirmaValida = Lx509.subjectPublicKeyRSA.verifyString(LtmpSignature, resp.signature);
    var LResul = false;
    //Validacion version Multibrowser 2.0.9
    if (! LfirmaValida) {
        //Validacion version Multibrowser 2.0.8
        LtmpSignature = encodedResp.replace(/"signature":"[^"]+"/, '"signature":""');
        LtmpSignature = LtmpSignature.replace(/"token":"[^"]+"/, '"token":""');
        LfirmaValida = Lx509.subjectPublicKeyRSA.verifyString(LtmpSignature, resp.signature);
        if (! LfirmaValida) {
            LResul = true;
        }
    }
    return LResul;
}

function procesarJSON(data, requerimiento, procesarResultado) {
    var encodedData = JSON.stringify(data);
    console.debug("Autentia request:", data);

    try {
        var ajaxResult = jQuery.ajax ({
            headers: { 'Content-type':'text/plain; charset=ISO8859_1'},
            type: 'POST',
            url: 'https://plugin.autentia.mb:7777/' + requerimiento + '/' + data.token,
            success: function(encodedResp) {
                var resp = JSON.parse(encodedResp);
                console.debug("Autentia response:", resp);
                if (exitoAutentia(resp)) {

                    if (FirmaInvalidaMB(resp, encodedResp)) {
                        resp = {
                            ParamsGet: {
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
            error: function() {
                procesarResultado({
                ParamsGet: {
                    ercText: 'Error de comunicacion con la componente Autentia.'
                    }
                });
            },
            complete: function() {},
            dataType: 'text',
            data: encodedData
        });

        console.log('AjaxResult', ajaxResult);
        if (ajaxResult.readyState !== 1) {
            procesarResultado({ParamsGet: {erc: -1, ercText: 'AjaxError: ' + ajaxResult.statusText}});
        }
    } catch (e){
        console.log("ERROR", e);
        procesarResultado({ParamsGet: {erc: -1, ercText: 'ERROR: ' + e.message}});
    }
}

var transaccion={};
var objPaquete=[];
var prmSet=[];
var token='';

function mensajeBloqueo(mensaje, ruedita) {
    var mensajeHtml = '<h5 style="font-family: Arial;">&nbsp;&nbsp;';
    var urigif = location.href.replace(/[^/]+$/, "img/ruedita.gif");
    if (ruedita === true) {
            mensajeHtml += '<img style="vertical-align: middle" src="' + urigif + '" />&nbsp;&nbsp;&nbsp;';
    }
    mensajeHtml += mensaje + '&nbsp;&nbsp;</h5>';

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
            top: '20%',
            left: '37%',
            right: '',
            width: '400px',
            "font-family": "Arial",
            "font-weight": "Bold",
            "font-size": "10px"
        }
    }
}

function plgAutentiaJS()
{
    console.debug('Plugin Loaded!');
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

function glosaAutentia(respuesta) {
   return decodeURIComponent(escape((
       respuesta.ParamsGet.ErcText ||
       respuesta.ParamsGet.ercText ||
       respuesta.ParamsGet.ErcDesc ||
       respuesta.ParamsGet.ercDesc
   )));
}

function exitoAutentia(respuesta) {
    return parseInt(valueOrDefault(function() {return respuesta.ParamsGet.erc;})) === 0;
}

function errorFatalAutentia(respuesta) {
    return valueOrDefault(function() {return respuesta.error;}, "") != "";
}

/* ProcesoFirmaPDF2
 *
 * clase que maneja el proceso de firma de multiples documentos segun protocolo universal
 * utilizado por DEC5 esta clase permite firmas masivas con token en cualquier aplicacion
 * ya que no depende de ningun servicio particular del DEC
 *
 * urlDocumentos: arreglo js con las URLs de los documentos a firmar
 * selectorCertificado: function que recibe un listado de certificados y una funcion para continuar.
 *                      1. realiza seleccion de uno de los certificados
 *                      2. invoka la funcion que continua con el proceso
 * notificarFinProceso: la funcion que se debe invokar para indicar el fin del proceso,
 *                      sea un exito o un error
 */
var ProcesoFirmaPDF2 = function (urlDocumentos, selectorCertificado, notificarFinProceso) {
    this.codDocumento = urlDocumentos;
    this.selectorCertificado = selectorCertificado;
    this.certificadoSeleccionado = null;
    this.ctxtId = "";
    this.resultados = [];
    this.notificarFinProceso = notificarFinProceso;
    this.documentoPorFirmar = 0;
    this.lastToken = "";
}

ProcesoFirmaPDF2.prototype.getToken = function() {
    this.lastToken = Math.random().toString();
    return this.lastToken;
}

ProcesoFirmaPDF2.prototype.procesarListadoCertificados = function(resultado) {
    if (exitoAutentia(resultado)) {
        this.selectorCertificado(JSON.parse(resultado.ParamsGet.certificates), this.procesarSeleccionCertificado.bind(this));
    }
    else {
        // fallo el listado de los certificados disponibles
        jQuery.unblockUI();
        this.notificarFinProceso(false, resultado);
    }
}

ProcesoFirmaPDF2.prototype.procesarSeleccionCertificado = function(resultado) {
    if (exitoAutentia(resultado)) {
        this.certificadoSeleccionado = resultado.ParamsGet;
        this.solicitarFirmaDocumento();
    }
    else {
        // el certificado no fue seleccionado
        jQuery.unblockUI();
        this.notificarFinProceso(false, resultado);
   }
}

ProcesoFirmaPDF2.prototype.solicitarFirmaDocumento = function() {
    if (this.certificadoSeleccionado === null) {
        // todavia no tenemos seleccionado un certificado
        jQuery.blockUI(mensajeBloqueo("Seleccionando el certificado del firmante...", true));

        var transaccion = {
            paquete: [{comando: "CryptoAPI.ListCerts"}],
            token: this.getToken()
        }

        procesarJSON(transaccion, 'CryptoAPI', this.procesarListadoCertificados.bind(this));
    }
    else {
        var nroDocumentos = this.codDocumento.length;
        if (this.documentoPorFirmar < nroDocumentos) {
            // tenemos documentos por firmar
            // bloqueamos la UI
            //$.blockUI(mensajeBloqueo('Procesando documento ' + (this.documentoPorFirmar + 1) + ' de ' + this.codDocumento.length, true));

            var transaccion = {
                paquete: [{
                    comando: "CryptoAPI.SignDocument",
                    urlDoc: this.codDocumento[this.documentoPorFirmar],
                    idCert: this.certificadoSeleccionado.idCert,
                    rut: this.certificadoSeleccionado.rut,
                    ctxtId: this.ctxtId
                }],
                token: this.getToken()
            }

            procesarJSON(transaccion, 'CryptoAPI', this.procesarResultadoFirma.bind(this, new Date().getTime()));
        } else {
            // terminamos de firmar todo sin problemas
            jQuery.unblockUI();
            this.notificarFinProceso(true);
        }
    }
}

ProcesoFirmaPDF2.prototype.procesarResultadoFirma = function(started, resultado) {
    this.resultados.push(resultado);
    this.documentoPorFirmar += 1;
    var ended = new Date().getTime();
    console.log("CryptoAPI.SignDocument: " , (ended - started) + " ms");
    if (exitoAutentia(resultado)) {
        // la firma del documento termino con exito, seguimos firmando
        if (this.ctxtId == "") {
            this.ctxtId = resultado.ParamsGet.ctxtId;
        }
        this.solicitarFirmaDocumento();
    } else {
        // hubo un error de firma de un documento,
        // se notifica el error e interrumpe el proceso
        jQuery.unblockUI();
        this.notificarFinProceso(false, resultado);
    }
}

/* ProcesoFirmaPDF
 *
 * clase que maneja el proceso de firma de multiples documentos segun
 * protocolo del DEC4
 */
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
        jQuery.blockUI(mensajeBloqueo('Procesando documento ' + (this.documentoPorFirmar + 1) + ' de ' + this.codDocumento.length, true));

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
    } else {
        // terminamos de firmar todo sin problemas
        jQuery.unblockUI();
        this.notificarFinProceso(true);
    }
}

ProcesoFirmaPDF.prototype.procesarResultadoFirma = function(started, resultado) {
    this.resultados.push(resultado);
    this.documentoPorFirmar += 1;
    var ended = new Date().getTime();
    console.log("pdfXSign: " , (ended - started) + " ms");
    if (exitoAutentia(resultado)) {
        // la firma del documento termino con exito, seguimos firmando
        if (this.ctxtId == "") {
            this.ctxtId = resultado.ParamsGet.ctxtId;
        }
        this.solicitarFirmaDocumento();
    } else {
        // hubo un error de firma de un documento,
        // se notifica el error e interrumpe el proceso
        jQuery.unblockUI();
        this.notificarFinProceso(false, resultado);
    }
}

plgAutentiaJS.prototype.firmarDocumentos = function(urlDocumentos, selectorCertificado, notificarFinProceso) {
    var procesoFirma = new ProcesoFirmaPDF2(urlDocumentos, selectorCertificado, notificarFinProceso);
    procesoFirma.solicitarFirmaDocumento();
};

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
    response = procesarJSON(transaccion, 'initSesion', function(response) {
        transaccion = {};
        objPaquete = [];
        prmSet = [];

        return excResp(response);
    });
};

plgAutentiaJS.prototype.IniciarSesionLogin = function(rut, token, excResp) {
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
    response = procesarJSON(transaccion, 'initSesion', function(response) {
        transaccion = {};
        objPaquete = [];
        prmSet = [];

        return excResp(response);
    });
}

plgAutentiaJS.prototype.Transaccion2 = function(trxName, entrada, salida, hookAutentia, token, excResp) {
    jQuery.blockUI(mensajeBloqueo('Ejecutando transaccion autentia.'));

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
        jQuery.unblockUI();
        return excResp(response);
    });

}

plgAutentiaJS.prototype.CerrarSesion = function(token) {
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
