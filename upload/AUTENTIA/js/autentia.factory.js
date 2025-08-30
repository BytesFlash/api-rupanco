var urlJS = '';
function autentiaFactory()
{
	var ua = navigator.userAgent.toLowerCase();
	var urlDsktp = 'http://200.0.156.150/pack_test_full_mbrowsertest/js/desktop/plugin.autentia.js';
	var urlMobile = 'http://200.0.156.150/pack_test_full_mbrowsertest/js/mobile/plugin.autentia.js';
	var isAndroid = ua.indexOf("android") > -1; //&& ua.indexOf("mobile");
	if(isAndroid) {
		urlJS = urlMobile;
	} else
	{
		urlJS = urlDsktp;
	}
	
}

autentiaFactory.prototype.OperacionAutentia = function(metodo, rutVerif, excResp) 
{
	$.when(
	    $.getScript( urlJS ),
	    $.Deferred(function( deferred ){
	        $( deferred.resolve );
	    })
	).done(function(){
		var objJS = new plgAutentiaJS();
		switch (metodo) {
			case "VerificaHuellaStandard":
				var entrada = {
					pRut:$("#rut").val()
				};
				var token = Math.random();
				var salida = ['Erc', 'ErcDesc', 'NroAudit'];
				objJS.Transaccion2('../FEDAUTENTIA/verificadatos', entrada, salida, true, token, function(response) {
					return excResp(response);
				});
				break;
			case "VerificaHuellaDEC4":
				var entrada = {
					Rut:rutVerif
				};
				var token = Math.random();
				var salida = ['Erc', 'ErcDesc', 'NroAudit'];
				objJS.Transaccion2('../DEC4/verificaDEC', entrada, salida, true, token, function(response) {
					return excResp(response);
				});
				break;
			default:
				return excResp(response);
		}	
	});	

}
