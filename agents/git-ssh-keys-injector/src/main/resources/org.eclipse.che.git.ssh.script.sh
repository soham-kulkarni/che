#
# Copyright (c) 2012-2017 Codenvy, S.A.
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Eclipse Public License v1.0
# which accompanies this distribution, and is available at
# http://www.eclipse.org/legal/epl-v10.html
#
# Contributors:
#   Codenvy, S.A. - initial API and implementation
#

script="/home/user/.ssh/git.sh"

token_suffix=$(if [ "$USER_TOKEN" != "dummy_token" ]; then echo &token=$USER_TOKEN; fi)
che_host=$(cat /etc/hosts | grep che-host | awk '{print $1;}')
api_url=$(if [ "$CHE_API" != "http://che-host:8080/wsmaster/api" ]; then echo $CHE_API; else echo $che_host:8080/api; fi)

echo '#!/bin/sh' > $script
echo 'host=$(echo "$(if [ "$1" == "-p" ]; then echo "$3" | sed -e '\'''s/git@//'\''; else echo "$1" | sed -e '\''s/git@//'\''; fi)")' >> $script
echo 'token_suffix='$token_suffix >> $script
echo 'che_host='$che_host >> $script
echo 'api_url='$api_url >> $script
echo 'ssh_key=$(echo "$(curl -s "$CHE_API/ssh/vcs/find?name=$host&token=$token"| grep -Po '\''"privateKey":.*?[^\\\]",'\'')" | sed -e "s/\"privateKey\":\"//")' >> $script
echo 'if [ -n "$ssh_key" ]' >> $script
echo 'then' >> $script
echo '    key_file=$(mktemp)' >> $script
echo '    echo "$ssh_key" > "$key_file"' >> $script
echo '    ssh -i "$key_file" "$@"' >> $script
echo '    rm "$key_file"' >> $script
echo 'else' >> $script
echo '    ssh "$@"' >> $script
echo 'fi' >> $script

chmod +x $script

git_committer_name=$(curl $api_url/preferences$token_suffix | grep -Po '"git.committer.name":.*?[^\\]",' | sed -e "s/\"git.committer.name\":\"//" | sed -e "s/\",//")
git_committer_email=$(curl $api_url/preferences$token_suffix | grep -Po '"git.committer.email":.*?[^\\]",' | sed -e "s/\"git.committer.email\":\"//" | sed -e "s/\",//")

echo 'export GIT_SSH='$script >> /home/user/.bashrc
echo 'export GIT_COMMITTER_NAME='$git_committer_name >> /home/user/.bashrc
echo 'export GIT_COMMITTER_EMAIL='$git_committer_email >> /home/user/.bashrc