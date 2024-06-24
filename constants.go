package binlog_modifier

const OPTION_NO_FOREIGN_KEY_CHECKS uint32 = 1 << 26

const (
	Q_FLAGS2_CODE = iota
	Q_SQL_MODE_CODE
	/*
		Q_CATALOG_CODE is catalog with end zero stored; it is used only by MySQL
		5.0.x where 0<=x<=3. We have to keep it to be able to replicate these
		old masters.
	*/
	Q_CATALOG_CODE
	Q_AUTO_INCREMENT
	Q_CHARSET_CODE
	Q_TIME_ZONE_CODE
	/*
		Q_CATALOG_NZ_CODE is catalog withOUT end zero stored; it is used by MySQL
		5.0.x where x>=4. Saves one byte in every Query_event in binlog,
		compared to Q_CATALOG_CODE. The reason we didn't simply re-use
		Q_CATALOG_CODE is that then a 5.0.3 slave of this 5.0.x (x>=4)
		master would crash (segfault etc) because it would expect a 0 when there
		is none.
	*/
	Q_CATALOG_NZ_CODE
	Q_LC_TIME_NAMES_CODE
	Q_CHARSET_DATABASE_CODE
	Q_TABLE_MAP_FOR_UPDATE_CODE
	Q_MASTER_DATA_WRITTEN_CODE
	Q_INVOKER
	/*
		Q_UPDATED_DB_NAMES status variable collects information of accessed
		databases i.e. the total number and the names to be propagated to the
		slave in order to facilitate the parallel applying of the Query events.
	*/
	Q_UPDATED_DB_NAMES
	Q_MICROSECONDS
	/*
	   A old (unused now) code for Query_log_event status similar to G_COMMIT_TS.
	*/
	Q_COMMIT_TS
	/*
	   A code for Query_log_event status, similar to G_COMMIT_TS2.
	*/
	Q_COMMIT_TS2
	/*
		The master connection @@session.explicit_defaults_for_timestamp which
		is recorded for queries, CREATE and ALTER table that is defined with
		a TIMESTAMP column, that are dependent on that feature.
		For pre-WL6292 master's the associated with this code value is zero.
	*/
	Q_EXPLICIT_DEFAULTS_FOR_TIMESTAMP
)
