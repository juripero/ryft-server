#!/bin/bash

getScriptDir() {
    dir=`dirname ${BASH_SOURCE}`
    cd $dir
    fullpathname=`pwd`
    cd - >/dev/null
    echo $fullpathname
}

scriptDir=`getScriptDir`
cd ${scriptDir}

# source common functions
#. ./functions
#setOSVariables

getRpmRelease() {
	echo $1 |\
	egrep -o "\-([0-9]+\.[0-9]+\.[0-9]+.*)_" |\
	sed 's/^\-//; s/_$//; s/\-/-/g'
}

#
returnCode=1

echo "# build_rpm.sh $@"

rpmRelease=`getRpmRelease $1`	
echo "# build_rpm rpm release=${rpmRelease}"

outfile=/tmp/build_deb_$$.log

echo
echo "alien --to-rpm --script $1 >$outfile"
alien --to-rpm --script $1 >$outfile
if [ $? == 0 ]
then
	echo "alien success"
        rpmfile=`cat $outfile|awk '{print $1}'`	

	echo "cksum ${rpmfile}"
	cksum ${rpmfile}

	export RPMREBUILD_TMPDIR=`mktemp -d`
	echo
	echo rpmrebuild  --change-files="sed -i 's|^%dir.*\"/\"||g;s|^%dir.*\"/usr/bin\"||g;s|^Version:.*|Version:14.32|g' $RPMREBUILD_TMPDIR/work/files.1" --define "_rpmfilename ryft-server-${rpmRelease}.x86_64.rpm" -p ${rpmfile} ">${outfile}.2"
	rpmrebuild  --change-files="sed -i 's|^%dir.*\"/\"||g;s|^%dir.*\"/usr/bin\"||g;' $RPMREBUILD_TMPDIR/work/files.1" --define "_rpmfilename ryft-server-${rpmRelease}.x86_64.rpm" -p ${rpmfile} >${outfile}.2


	if [ $? == 0 ]
	then
		echo "rpmrebuild success file=" `cat ${outfile}.2`
		rpmfile=`awk '{print $2}' ${outfile}.2`
	
		echo "rpmrebuild --define \"_rpmfilename ryft-server-${rpmRelease}_x86_64.rpm\" -p ${rpmfile} >${outfile}.3"
		rpmrebuild --define "_rpmfilename ryft-server-${rpmRelease}_x86_64.rpm" -p ${rpmfile} >${outfile}.3
		echo "rpmrebuild success file=" `cat ${outfile}.2`
		rpmfile2=`awk '{print $2}' ${outfile}.2`
	
		echo "cksum ${rpmfile2}"
		cksum ${rpmfile2}

		echo "cp ${rpmfile2} /opt/debian/"
		cp ${rpmfile2} /opt/debian/

		echo "rm ${scriptDir}/*.rpm"
		rm ${scriptDir}/*.rpm

		returnCode=0
	else
		echo "rpmrebuild failure"
		returnCode=3
	fi
else
	echo "alien failure"
	returnCode=2
fi
exit $returnCode
