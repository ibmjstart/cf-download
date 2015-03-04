
export PYTHONPATH=$HOME/.bp/lib
$HOME/.bp/bin/rewrite "$HOME/httpd/conf"
$HOME/.bp/bin/rewrite "$HOME/php/etc"
source $HOME/.env
$HOME/.bp/bin/start
