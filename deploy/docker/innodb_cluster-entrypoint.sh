#!/bin/bash
set -e

echo "This is start of the container"
# if command starts with an option, save them as CMD arguments
if [ "${1:0:1}" = '-' ]; then
        ARGS="$@"
fi

MYSQL_ROOT_PASSWORD="passw0rd"
MYSQL_PORT="3306"
MYSQL_USER="root"
CLUSTER_NAME="testcluster"

echo "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"
env
echo "==================================="

# If we're setting up a router 
if [ "$NODE_TYPE" = 'router' ]; then

       echo "Setting up a new router instance..."

       echo "This is the MYSQL_HOST:$MYSQL_HOST"

       # we need to ensure that they've specified a boostrap URI 
        if [ -z "$PODPREFIX" -a -z "$MYSQL_PASSWORD" ]; then
                echo >&2 "error: You must specify a value for MYSQL_HOST and MYSQL_PASSWORD (MYSQL_USER=root is the default) when setting up a router"
                exit 1
        fi

        # We'll use the hostname as the router instance name
	HOSTNAME=$(hostname).mysql
        # first we need to see if the cluster metadata already exists 
	set +e
	for i in 0 1 2
	do 
		metadata_exists=$(mysqlsh --uri="$MYSQL_USER"@"${PODPREFIX}-${i}.mysql":"$MYSQL_PORT" -p"$MYSQL_ROOT_PASSWORD" --no-wizard --js -i -e "dba.getCluster( '${CLUSTER_NAME}' )" 2>&1 | grep "<Cluster:$CLUSTER_NAME>")

		if [ ! -z "$metadata_exists" ]; then
			MYSQL_HOST=$(mysql --no-defaults -h "${PODPREFIX}-${i}.mysql" -P"$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_ROOT_PASSWORD" -nsLNE -e "select member_host from performance_schema.replication_group_members where member_state='ONLINE' and member_id=(IF((select @grpm:=variable_value from performance_schema.global_status where variable_name='group_replication_primary_member') = '', member_id, @grpm)) limit 1" 2>/dev/null | grep -v '*')
			break
		fi
	done
        set -e
	echo "Tested cluster"

        if [ -z "$metadata_exists" ]; then
		echo "No cluster was found!"
		# Then let's create the innodb cluster metadata 
		output=$(mysqlsh --uri="$MYSQL_USER"@"${PODPREFIX}-0.mysql":"$MYSQL_PORT" -p"$MYSQL_ROOT_PASSWORD" --no-wizard --js -i -e "dba.createCluster('${CLUSTER_NAME}', {adoptFromGR: true})")
		echo "Created cluster"
		MYSQL_HOST=${PODPREFIX}-0.mysql
	fi
	echo "MySQL host used to bootstrap router is:$MYSQL_HOST"
        output=$(echo "$MYSQL_ROOT_PASSWORD" | mysqlrouter --bootstrap="$MYSQL_USER"@"$MYSQL_HOST":"$MYSQL_PORT" --user=mysql --name "$HOSTNAME" --force)

        if [ ! "$?" = "0" ]; then
		echo "error: could not bootstrap router:"
		echo "$output"
		exit 1
	fi
       
        # bug (?) in Router 2.1.3 didn't set file permissions based on --user value 
	chown -R mysql:mysql "/var/lib/mysqlrouter"

        # now that we've bootstrapped the setup, let's start the process
        CMD="mysqlrouter --user=mysql"

# Let's setup a mysql server instance normally 
else
	#Check if cluster is already created. This is required when pod is being restarted.
	set +e
	for i in 0 1 2
	do 
	metadata_exists=$(mysqlsh --uri="$MYSQL_USER"@"${PODPREFIX}-${i}.mysql":"$MYSQL_PORT" -p"$MYSQL_ROOT_PASSWORD" --no-wizard --js -i -e "dba.getCluster( '${CLUSTER_NAME}' )" 2>&1 | grep "<Cluster:$CLUSTER_NAME>")

	if [ ! -z "$metadata_exists" ]; then
		echo "Find exist cluster!"
		break
	fi
	done
        set -e

	#Get sequence number of the pod and use it to set server ID
	[[ `hostname` =~ -([0-9]+)$ ]] || exit 1
	ordinal=${BASH_REMATCH[1]}
	SERVER_ID=$((100 + $ordinal))
	
	if [ -z "$metadata_exists" ]; then  #no cluster is found
		if [[ $ordinal -eq 0 ]]; then
			echo "Set to bootstrap"
			BOOTSTRAP=1
		else
			#Verify pod-0 has been created. Statefulset dose not make sure pod-0 is created before the others when pods are restarted.
			no_pod0=0
			set +e
			output=$(mysqlsh --uri="$MYSQL_USER"@"${PODPREFIX}-0.mysql":"$MYSQL_PORT" -p"$MYSQL_ROOT_PASSWORD" --no-wizard --sql -i -e "SELECT 1;" 2>&1 > /dev/null) || no_pod0=$?
			set -e
			if [ ! "$no_pod0" = "0" ]; then
				echo "error: no primary pod is found!"
				exit 1
			fi
	        fi
	fi
	
	GROUP_NAME="925deea1-75db-4e92-86a9-00259c8d8078"
	GROUP_SEEDS="${PODPREFIX}-0.mysql:6606,${PODPREFIX}-1.mysql:6606,${PODPREFIX}-2.mysql:6606"

	#use a static group name
	CMD="mysqld"

        # We'll use this variable to manage the mysqld args 
        MYSQLD_ARGS="--server_id=$SERVER_ID"

	# Test we're able to startup without errors. We redirect stdout to /dev/null so
	# only the error messages are left.
	result=0
	output=$("$CMD" --verbose --help $MYSQLD_ARGS 2>&1 > /dev/null) || result=$?
	if [ ! "$result" = "0" ]; then
		echo >&2 'error: could not run mysql. This could be caused by a misconfigured my.cnf'
		echo >&2 "$output"
		exit 1
	fi

	# Get config
	DATADIR="$("$CMD" --verbose --help --log-bin-index=/tmp/tmp.index $MYSQLD_ARGS 2>/dev/null | awk '$1 == "datadir" { print $2; exit }')"

	GR_ARGS="--plugin-load=group_replication.so --group_replication_start_on_boot=ON "

	# if we're bootstrapping a new group then let's just generate a new group_name / UUID	
	if [ ! -z "$BOOTSTRAP" ]; then
		# Let's not blindly bootstrap the cluster if the datadir already exists
		# In that case we've likely restarted an existing container
		#if [ ! -d "$DATADIR/mysql" ]; then
			GR_ARGS="$GR_ARGS --group_replication_bootstrap_group=ON "

			# Let's generate a UUID if one hasn't been specified 
			[ -z "$GROUP_NAME" ] && GROUP_NAME="925deea1-75db-4e92-86a9-00259c8d8078"

			#Let's persist the group_name since the env variable is not set
			#This will allow for restarting the container w/o bootstrapping a new/second cluster
			echo "loose-group-replication-group-name=$GROUP_NAME" >> /etc/mysql/my.cnf
		
			echo >&1 "info: Bootstrapping new Group Replication cluster using --group_replication_group_name=$GROUP_NAME"
			echo >&1 "  You will need to specify GROUP_NAME=$GROUP_NAME if you want to add another node to this cluster"
		#fi
        else
		echo >&1 "info: attempting to join the $GROUP_NAME group using $GROUP_SEEDS as seeds"
        	GR_ARGS="$GR_ARGS --group_replication_group_name=$GROUP_NAME"
                GR_ARGS="$GR_ARGS --group_replication_group_seeds=$GROUP_SEEDS"
	fi

        # You can use --hostname=<hostname> for each container or use the auto-generated one; 
        # we'll need to use the hostname for group_replication_local_address
        HOSTNAME=$(hostname)

	GR_ARGS="$GR_ARGS --group_replication_local_address=$HOSTNAME.mysql.default.svc.cluster.local:6606 --report-host=$HOSTNAME.mysql"

	if [ ! -d "$DATADIR/mysql" ]; then
		if [ -z "$MYSQL_ROOT_PASSWORD" -a -z "$MYSQL_ALLOW_EMPTY_PASSWORD" -a -z "$MYSQL_RANDOM_ROOT_PASSWORD" ]; then
			echo >&2 'error: database is uninitialized and password option is not specified '
			echo >&2 '  You need to specify one of MYSQL_ROOT_PASSWORD, MYSQL_ALLOW_EMPTY_PASSWORD and MYSQL_RANDOM_ROOT_PASSWORD'
			exit 1
		fi
		mkdir -p "$DATADIR"
		chown -R mysql:mysql "$DATADIR"

		echo "Initializing database"
		"$CMD" --initialize-insecure=on $MYSQLD_ARGS
		echo "Database initialized"

		"$CMD" --skip-networking $MYSQLD_ARGS &
		pid="$!"

		mysql=( mysql --protocol=socket -uroot )

		for i in {30..0}; do
			if echo "SELECT 1" | "${mysql[@]}" &> /dev/null; then
				break
			fi
			echo "MySQL init process in progress..."
			sleep 1
		done
		if [ "$i" = 0 ]; then
			echo >&2 "MySQL init process failed."
			exit 1
		fi

		mysql_tzinfo_to_sql /usr/share/zoneinfo | "${mysql[@]}" mysql
		
		if [ ! -z "$MYSQL_RANDOM_ROOT_PASSWORD" ]; then
			MYSQL_ROOT_PASSWORD="$(pwmake 128)"
			echo "GENERATED ROOT PASSWORD: $MYSQL_ROOT_PASSWORD"
		fi
		"${mysql[@]}" <<-EOSQL
			SET @@SESSION.SQL_LOG_BIN=0;
			DELETE FROM mysql.user WHERE user NOT IN ('mysql.session', 'mysql.sys', 'mysqlxsys') OR host NOT IN ('localhost');
			CREATE USER 'root'@'%' IDENTIFIED BY '${MYSQL_ROOT_PASSWORD}' ;
			GRANT ALL ON *.* TO 'root'@'%' WITH GRANT OPTION ;
			DROP DATABASE IF EXISTS test ;
			FLUSH PRIVILEGES ;
		EOSQL
		if [ ! -z "$MYSQL_ROOT_PASSWORD" ]; then
			mysql+=( -p"$MYSQL_ROOT_PASSWORD" )
		fi

		if [ "$MYSQL_DATABASE" ]; then
			echo "CREATE DATABASE IF NOT EXISTS \`$MYSQL_DATABASE\` ;" | "${mysql[@]}"
			mysql+=( "$MYSQL_DATABASE" )
		fi

		if [ "$MYSQL_USER" -a "$MYSQL_PASSWORD" ]; then
			echo "CREATE USER '${MYSQL_USER}'@'%' IDENTIFIED BY '${MYSQL_PASSWORD}' ;" | "${mysql[@]}"

			if [ "$MYSQL_DATABASE" ]; then
				echo "GRANT ALL ON ${MYSQL_DATABASE}.* TO '${MYSQL_USER}'@'%' ;" | "${mysql[@]}"
			fi

			echo "FLUSH PRIVILEGES ;" | "${mysql[@]}"
		fi

		echo 

               for f in /docker-entrypoint-initdb.d/*; do
	                case "$f" in
		        	   *.sh)  echo "$0: running $f"; . "$f" ;;
				   *.sql) echo "$0: running $f"; "${mysql[@]}" < "$f" && echo;; 
				   *)     echo "$0: ignoring $f" ;;
		        esac
		   echo
	        done

		if [ ! -z "$MYSQL_ONETIME_PASSWORD" ]; then
			"${mysql[@]}" <<-EOSQL
				ALTER USER 'root'@'%' PASSWORD EXPIRE;
			EOSQL
		fi


       		echo "RESET MASTER ;" | "${mysql[@]}"

                # lastly we need to setup the recovery channel with a valid username/password
                echo "CHANGE MASTER TO MASTER_USER='root', MASTER_PASSWORD='${MYSQL_ROOT_PASSWORD}' FOR CHANNEL 'group_replication_recovery' ;" | "${mysql[@]}"

		if ! kill -s TERM "$pid" || ! wait "$pid"; then
			echo >&2 "MySQL init process failed."
			exit 1
		fi

		echo
		echo "MySQL init process done. Ready for start up."
		echo
	fi

	chown -R mysql:mysql "$DATADIR"

        CMD="mysqld $ARGS $GR_ARGS $MYSQLD_ARGS"
fi
echo "This is the end!"
echo $CMD
exec $CMD

