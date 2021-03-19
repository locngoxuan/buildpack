Name:		buildpack
Version:	%{BUILD_VERSION}
Release:    %{BUILD_ID}.%{BUILD_OS}

Summary:	Independent build tool
Group:		Development/Tools
License:	GPLv3+
Source0:	buildpack-%{version}.tar.gz

%description
Run buildpack --help for more detail.

%prep
%setup -q

%install
rm -rf $RPM_BUILD_ROOT
install -d $RPM_BUILD_ROOT/etc/buildpack
install -d $RPM_BUILD_ROOT/etc/buildpack/plugins
install -d $RPM_BUILD_ROOT/etc/buildpack/plugins/builder
install -d $RPM_BUILD_ROOT/etc/buildpack/plugins/publisher
install buildpack $RPM_BUILD_ROOT/etc/buildpack/buildpack

%clean
rm -rf $RPM_BUILD_ROOT

%files
%defattr(-,root,root,-)
/etc/buildpack/buildpack
/etc/buildpack/plugins
/etc/buildpack/plugins/builder
/etc/buildpack/plugins/publisher
%doc

%pre
rm -rf /etc/buildpack
rm -rf /usr/bin/buildpack

%post
echo "---------------------------------------------------------------------------------------------";
echo "";
echo " Package information";
echo " Version: %{BUILD_VERSION}";
echo " Product Version: %{BUILD_VERSION}";
echo " Home Directory /etc/buildpack";
echo " Plugin Directory /etc/buildpack/plugins";
echo "";
echo "---------------------------------------------------------------------------------------------";
echo;

# Now create the links for buildpack to use ...
chmod 755 /etc/buildpack/buildpack
ln -s /etc/buildpack/buildpack /usr/bin/buildpack

echo;
echo "---------------------------------------------------------------------------------------------";
echo " Install successful ";
echo "---------------------------------------------------------------------------------------------";
