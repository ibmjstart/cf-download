
<?php
/**
 * VCAP CONFIG
 * The base configurations of the WordPress.
 *
 * This file has the following configurations: MySQL settings, Table Prefix,
 * Secret Keys, WordPress Language, and ABSPATH. You can find more information
 * by visiting {@link http://codex.wordpress.org/Editing_wp-config.php Editing
 * wp-config.php} Codex page. You can get the MySQL settings from your web host.
 *
 * This file is used by the wp-config.php creation script during the
 * installation. You don't have to use the web site, you can just copy this file
 * to "wp-config.php" and fill in the values.
 *
 * @package WordPress
 */

// ** MySQL settings - You can get this info from your web host ** //
/** The name of the database for WordPress */

$vcap = getenv("VCAP_SERVICES");
$data = json_decode($vcap, true);
$creds = $data['cleardb'][0]['credentials'];
define('DB_NAME', $creds['name']);

/** MySQL database username */
define('DB_USER', $creds['username']);

/** MySQL database password */
define('DB_PASSWORD', $creds['password']);

/** MySQL hostname */
define('DB_HOST', $creds['hostname']);

/** Database Charset to use in creating database tables. */
define('DB_CHARSET', 'utf8');

/** The Database Collate type. Don't change this if in doubt. */
define('DB_COLLATE', '');

define( 'AUTOMATIC_UPDATER_DISABLED', true );

/** Allow DB Repair*/
define('WP_ALLOW_REPAIR', true);

/**#@-*/

/**
 * WordPress Database Table prefix.
 *
 * You can have multiple installations in one database if you give each a unique
 * prefix. Only numbers, letters, and underscores please!
 */
$table_prefix  = 'bluemix_0_5_wp_';

/**
 * WordPress Localized Language, defaults to English.
 *
 * Change this to localize WordPress. A corresponding MO file for the chosen
 * language must be installed to wp-content/languages. For example, install
 * de_DE.mo to wp-content/languages and set WPLANG to 'de_DE' to enable German
 * language support.
 */
define('WPLANG', '');

/**
 * For developers: WordPress debugging mode.
 *
 * Change this to true to enable the display of notices during development.
 * It is strongly recommended that plugin and theme developers use WP_DEBUG
 * in their development environments.
 */
define('WP_DEBUG', false);

/* That's all, stop editing! Happy blogging. */

/** Absolute path to the WordPress directory. */
if ( !defined('ABSPATH') )
	define('ABSPATH', dirname(__FILE__) . '/');

/** Sets up WordPress vars and included files. */

define( 'WP_DEFAULT_THEME', 'twentyfourteen' );

require_once(ABSPATH . 'wp-settings.php');

require_once ABSPATH.'/wp-admin/includes/plugin.php';
/*
 * 		Set the default plugins that we're going to use
 * 		for WordPress on Bluemix.
 */
if(!get_option('default_plugins_activated')){
	update_option('default_plugins_activated', "1");
	activate_plugin( 'wp-bluemix-objectstorage/objectstorage.php' );
	activate_plugin( 'stops-core-theme-and-plugin-updates/main.php');
	activate_plugin( 'wp-bluemix-sendgrid/bluemix-sendgrid.php');
	activate_plugin( 'sendgrid-email-delivery-simplified/wpsendgrid.php');
	//activate_plugin( 'wp-super-cache/wp-cache.php');

	if(!get_option('_disable_updates')){
  		update_option('_disable_updates', array(
    		'all' => '1',
  		));
	}

	if(!get_option('disable_updates_blocked')){
	  	update_option('disable_updates_blocked', array());
	}
}

/**#@+
 * Authentication Unique Keys and Salts.
 *
 * Change these to different unique phrases!
 * You can generate these using the {@link https://api.wordpress.org/secret-key/1.1/salt/ WordPress.org secret-key service}
 * You can change these at any point in time to invalidate all existing cookies. This will force all users to have to log in again.
 *
 * @since 2.6.0
 */
//This is at least one method of generating the necessary salt as needed randomly, so that every user of the boilerplate gets a new one.

define('AUTH_KEY',         get_option('auth_key'));
define('SECURE_AUTH_KEY',  get_option('secure_auth_key'));
define('LOGGED_IN_KEY',    get_option('logged_in_key'));
define('NONCE_KEY',        get_option('nonce_key'));
define('AUTH_SALT',        get_option('auth_salt'));
define('SECURE_AUTH_SALT', get_option('secure_auth_salt'));
define('LOGGED_IN_SALT',   get_option('logged_in_salt'));
define('NONCE_SALT',       get_option('nonce_salt'));

