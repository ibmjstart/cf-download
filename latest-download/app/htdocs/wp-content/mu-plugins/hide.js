
/*
*	Insert warning to the top of the "add new plugin" and "add new theme" pages because they'll just disappear when you restart your app.
*/

jQuery(function() {	
	jQuery('#wpbody-content > div.wrap > h2').after('<div style="color:#990033;display:block;text-align:center">Warning: Changes made to your WordPress configuration will be lost. See our <a href="https://www.ng.bluemix.net/docs/#starters/wordpress/index.html">docs</a> for information on how to customize WordPress on Bluemix.</div>');
});

