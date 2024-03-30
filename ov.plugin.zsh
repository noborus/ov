# make sure you execute this *after* asdf or other version managers are loaded

if (( $+commands[ov] )); then
  eval "$(ov --completion zsh)"
fi