var constXMLSignature = "<Signature xmlns='http://www.w3.org/2000/09/xmldsig#' Id=''> \
	<SignedInfo> \
		<CanonicalizationMethod Algorithm='http://www.w3.org/TR/2001/REC-xml-c14n-20010315'/> \
		<SignatureMethod Algorithm='http://www.w3.org/2000/09/xmldsig#rsa-sha1'/> \
		<Reference URI='#'> \
			<DigestMethod Algorithm='http://www.w3.org/2000/09/xmldsig#sha1'/> \
			<DigestValue></DigestValue> \
		</Reference> \
		<Reference URI='#'> \
			<DigestMethod Algorithm='http://www.w3.org/2000/09/xmldsig#sha1'/> \
			<DigestValue></DigestValue> \
		</Reference> \
		<Reference URI='#'> \
			<DigestMethod Algorithm='http://www.w3.org/2000/09/xmldsig#sha1'/> \
			<DigestValue></DigestValue> \
		</Reference> \
	</SignedInfo> \
	<SignatureValue></SignatureValue> \
	<KeyInfo> \
		<KeyValue> \
			<RSAKeyValue> \
				<Modulus></Modulus> \
				<Exponent></Exponent> \
			</RSAKeyValue> \
		</KeyValue> \
		<X509Data> \
			<X509Certificate></X509Certificate> \
		</X509Data> \
	</KeyInfo> \
	<Object> \
		<SignatureProperties> \
			<SignatureProperty Id='' Target='#'> \
				<BDSRequest xmlns='http://acepta.com/bds'> \
					<BDSRequestId></BDSRequestId> \
				</BDSRequest> \
			</SignatureProperty> \
		</SignatureProperties> \
		<AuditoriaAutentia xmlns='http://www.autentia.cl/auditoria' Id=''></AuditoriaAutentia> \
	</Object> \
</Signature>";

var myXML = "";
/*
var kernel32 = new ActiveXObject('DynamicWrapper');
kernel32.Register("kernel32.dll", "OutputDebugStringA", "I=s", "F=s", "R=u");
kernel32.Register("kernel32.dll", "QueryDosDeviceA", "I=uuu", "F=s", "R=u");
*/

var documentID = "_folioDocumento_";
var signatureID = Math.random().toString();
var signaturePropertyID = Math.random().toString();
var BDSRequestID = Math.random().toString();
var LAutentiaID = Math.random().toString();
var uriKey = Math.random().toString();

function GetYYMM() 
{
	var date = new Date();
	var mm = ("0" + (date.getMonth() + 1)).slice(-2);
	var yy = date.getFullYear().toString().substr(2,2);
	
	return yy + mm;
};

function getNroDocumento() {
//	kernel32.OutputDebugStringA("** DEBUG JS ** getNroDocumento: ");
	var nroDoc;
	try {
		nodo = myXML.getElementsByTagName('Document/Content/documentos/documento')[0];
		nroDoc = nodo.getAttribute("id");
//		kernel32.OutputDebugStringA("** DEBUG JS ** nroDoc: " + nroDoc);		
	} catch (e) {
//		kernel32.OutputDebugStringA('ERROR: ' + e.message);
		nroDoc = -1;
	}

	return nroDoc;	
}

function getNodeValue(args)
{
	var argsTmp = args.split("|");
	var xml = argsTmp[0];
	var xpath = argsTmp[1];
	var nodo; 

	try {
		var xmlDoc = new ActiveXObject("Microsoft.XMLDOM");
		xmlDoc.async = "false";
		xmlDoc.loadXML(xml);

//		kernel32.OutputDebugStringA("** DEBUG JS ** xpath: " + xpath);
		nodo = xmlDoc.getElementsByTagName(xpath)[0].childNodes[0];
//		kernel32.OutputDebugStringA("** DEBUG JS ** getNodeValue " + nodo.nodeValue);
	} catch (e) {
//		kernel32.OutputDebugStringA('ERROR: ' + e.message);
	}

	return nodo;
}

//function CalcSHA1(data)
//{
//	return "AbCdEfG";
//}

function dwnld_xml(argsFromTrx) {
	var argsTrx = argsFromTrx.split("|");
	var nroAudit = argsTrx[0];
	var urlXML = argsTrx[1];
	//var outXML = '';

//	kernel32.OutputDebugStringA("** DEBUG JS ** Inicio function dwnld_xml_MultiDocs()...; ");

	var outXML = {
		"numeroSolicitud" :"3908601",
		"cantidadFirmada" :3,
		"documentos": [{
		    "idDocumento": 123225,
			"documento": "CONTRATO",
			"rut": "15984151-0",
			"nombre": "Daniella",
			"estado": "FIRMADO",
			"URI": "ID_DOC_1"
		},
		{
		    "idDocumento": 1232256,
			"documento": "CONTRATO",
			"rut": "15984151-0",
			"nombre": "Daniella",
			"estado": "FIRMADO",
			"URI": "ID_DOC_2"
		},
		{
		    "idDocumento": 1232257,
			"documento": "CONTRATO",
			"rut": "15984151-0",
			"nombre": "Daniella",
			"estado": "FIRMADO",
			"URI": "ID_DOC_3"
		}]		
	}

	return (JSON.stringify(outXML));
}