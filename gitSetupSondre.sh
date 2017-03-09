echo "Unset global Git users"
git config --global --unset-all user.name
git config --global --unset-all user.email 


echo "Set Anders as local repo user"
git config user.name "Sondre Kongsgård"
git config user.email kingsondis@gmail.com


echo "set Meld as default mergetool"
git config --global merge.tool meld

echo "set Meld as default difftool"
git config --global diff.guitool meld
git config --global diff.tool meld
git config --global difftool.promt false


