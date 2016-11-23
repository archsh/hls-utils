/*
   hls-get

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/
package main

import (
	"flag"
	"os"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/archsh/hlsutils/helpers/logging"
	"github.com/BurntSushi/toml"
)

const VERSION = "0.9.9"

var logging_config = logging.LoggingConfig{Format:logging.DEFAULT_FORMAT, Level:"DEBUG"}

type RedisConfig struct {
	Host string
	Port int
	Db int
	Password string
	Key string
}

type MySQLConfig struct {
	Host string
	Port int
	Username string
	Password string
	Db string
	Table string
}

type Configuration struct {
	Output          string
	Path_Rewrite    string
	Segment_Rewrite string
	User_Agent      string
	Log_File        string
	Log_Level       string
	Retries         int
	Skip            bool
	Mode            string
	Redirect        string
	Concurrent      int
	Timeout         int
	Total           int64
	Redis           RedisConfig
	MySQL           MySQLConfig
}

func Usage() {
	guide := `
Scenarios:
  (1) Simple mode: download one or multiple URL without DB support.
  (2) Redis support: download multiple URL via REDIS LIST.
  (3) MySQL support: download multiple URL via MySQL DB Table.

Usage:
  hls-get [OPTIONS,...] [URL1,URL2,...]

Options:
`
	os.Stdout.Write([]byte(guide))
	flag.PrintDefaults()
}

/***********************************************************************************************************************
 * MAIN ENTRY
 *
 */
func main() {
	//c 'config'    - [STRING] Config file instead of the following parameters. Default empty.
	var config string
	flag.StringVar(&config, "c", "", "Use a config file instead of the other parameters. Default empty.")
	//O  'output'     - [STRING] Output directory. Default '.'.
	var output string
	flag.StringVar(&output, "O", ".", "Output directory.")
	//PR 'path_rewrite'    - [STRING] Rewrite output path method. Default empty means simple copy.
	var path_rewrite string
	flag.StringVar(&path_rewrite, "PR", "", "Rewrite output path method. Empty means simple copy.")
	//SR 'segment_rewrite'     - [STRING] Rewrite segment name method. Default empty means simple copy.
	var segment_rewrite string
	flag.StringVar(&segment_rewrite, "SR", "", "Rewrite segment name method. Empty means simple copy.")
	//UA 'user_agent'    - [STRING] UserAgent. Default is 'hls-get' with version num.
	var user_agent string
	flag.StringVar(&user_agent, "UA", "hls-get v" + VERSION, "UserAgent.")
	//L  'log'   - [STRING] Logging output file. Default 'stdout'.
	var log_file string
	flag.StringVar(&log_file, "L", "", "Logging output file. Default 'stdout'.")
	//LV 'loglevel' - [STRING] Log level. Default 'INFO'.
	var log_level string
	flag.StringVar(&log_level, "LV", "INFO", "Logging level. Default 'INFO'.")
	//R  'retry' - [INTEGER] Retry times if download fails.
	var retries int
	flag.IntVar(&retries, "R", 0, "Retry times if download fails.")
	//S  'skip'  - [BOOL] Skip if exists.
	var skip bool
	flag.BoolVar(&skip, "S", false, "Skip if exists.")
	//M  'mode'  - [STRING] Source mode: redis, mysql. Default empty means source via command args.
	var mode string
	flag.StringVar(&mode, "M", "", "Source mode: redis, mysql. Empty means source via command args.")
	//RD 'redirect'   - [STRING] Redirect server request.
	var redirect string
	flag.StringVar(&redirect, "RR", "", "Redirect server request.")
	//C  'concurrent' - [INTEGER] Concurrent tasks.
	var concurrent int
	flag.IntVar(&concurrent, "C", 5, "Concurrent tasks.")
	//TO 'timeout'    - [INTEGER] Request timeout in seconds.
	var timeout int
	flag.IntVar(&timeout, "TO", 20, "Request timeout in seconds.")
	//TT 'total'      - [INTEGER] Total download links.
	var total int64
	flag.Int64Var(&total, "TT", 0, "Total download links.")
	//
	//RH 'redis_host'  - [STRING] Redis host.
	var redis_host string
	flag.StringVar(&redis_host, "RH", "localhost", "Redis host.")
	//RP 'redis_port'  - [INTEGER] Redis port.
	var redis_port int
	flag.IntVar(&redis_port, "RP", 6379, "Redis port.")
	//RD 'redis_db'    - [INTEGER] Redis db num.
	var redis_db int
	flag.IntVar(&redis_db, "RD", 0, "Redis db num.")
	//RW 'redis_password'  - [STRING] Redis password.
	var redis_password string
	flag.StringVar(&redis_password, "RW", "", "Redis password.")
	//RK 'redis_key'   - [STRING] List key name in redis.
	var redis_key string
	flag.StringVar(&redis_key, "RK", "HLSGET_DOWNLOADS", "List key name in redis.")
	//RU 'redis_url'   - [STRING] ${redis_host}:${redis_port}/${redis_db}/${redis_key}
	//var redis_url string
	//flag.StringVar(&redis_url, "RU", "", "${redis_host}:${redis_port}/${redis_db}/${redis_key}")
	//
	//MH 'mysql_host'  - [STRING] MySQL host.
	var mysql_host string
	flag.StringVar(&mysql_host, "MH", "localhost", "MySQL host.")
	//MP 'mysql_port'  - [INTEGER] MySQL port.
	var mysql_port int
	flag.IntVar(&mysql_port, "MP", 3306, "MySQL port.")
	//MN 'mysql_username' - [STRING] MySQL username.
	var mysql_username string
	flag.StringVar(&mysql_username, "MN", "root", "MySQL username.")
	//MW 'mysql_password' - [STRING] MySQL password.
	var mysql_password string
	flag.StringVar(&mysql_password, "MW", "", "MySQL password.")
	//MD 'mysql_db'       - [STRING] MySQL database.
	var mysql_db string
	flag.StringVar(&mysql_db, "MD", "hlsgetdb", "MySQL database.")
	//MT 'mysql_table'    - [STRING] MySQL table.
	var mysql_table string
	flag.StringVar(&mysql_table, "MT", "hlsget_downloads", "MySQL table.")
	var mysql_show_schema bool
	flag.BoolVar(&mysql_show_schema, "MS", false, "Only show MySQL table schema. Will not do any furthur action if this is set.")
	//MU 'mysql_url'      - [STRING] ${mysql_username}:${mysql_password}@${mysql_host}:${mysql_port}/${mysql_db}/${mysql_table}
	//var mysql_url string
	//flag.StringVar(&mysql_url, "MU", "", "${mysql_username}:${mysql_password}@${mysql_host}:${mysql_port}/${mysql_db}/${mysql_table}")

	flag.Parse()

	os.Stderr.Write([]byte(fmt.Sprintf("hls-get v%v - HTTP Live Streaming (HLS) Downloader.\n", VERSION)))
	os.Stderr.Write([]byte("Copyright (C) 2015 Mingcai SHEN <archsh@gmail.com>. Licensed for use under the GNU GPL version 3.\n"))
	if mysql_show_schema {
		ShowMySQLSchema()
		os.Exit(0)
	}
	if config != "" {
		cfg := new(Configuration)
		if _, e := toml.DecodeFile(config, cfg); nil != e {
			os.Stderr.Write([]byte(fmt.Sprintf("Load config<%s> failed: %s.\n", config, e)))
			os.Exit(1)
		}else{
			os.Stderr.Write([]byte(fmt.Sprintf("Loaded config from <%s> .\n", config)))
			output = cfg.Output
			path_rewrite = cfg.Path_Rewrite
			segment_rewrite = cfg.Segment_Rewrite
			if cfg.User_Agent != "" {
				user_agent = cfg.User_Agent
			}
			log_file = cfg.Log_File
			if cfg.Log_Level != "" {
				log_level = cfg.Log_Level
			}
			retries = cfg.Retries
			skip = cfg.Skip
			mode = cfg.Mode
			redirect = cfg.Redirect
			concurrent = cfg.Concurrent
			timeout = cfg.Timeout
			total = cfg.Total
			redis_host = cfg.Redis.Host
			redis_port = cfg.Redis.Port
			redis_db = cfg.Redis.Db
			redis_password = cfg.Redis.Password
			redis_key = cfg.Redis.Key
			mysql_host = cfg.MySQL.Host
			mysql_port = cfg.MySQL.Port
			mysql_username = cfg.MySQL.Username
			mysql_password = cfg.MySQL.Password
			mysql_db = cfg.MySQL.Password
			mysql_table = cfg.MySQL.Table
			//os.Stderr.Write([]byte(fmt.Sprintf("Loaded config: %+v .\n", cfg)))
		}
	}
	logging_config.Filename = log_file
	logging_config.Level = log_level
	if log_file != "" {
		logging.InitializeLogging(&logging_config, false, logging_config.Level)
	}else{
		logging.InitializeLogging(&logging_config, true, logging_config.Level)
	}
	defer logging.DeinitializeLogging()
	path_rewriter := NewPathRewriter(path_rewrite)
	segment_rewriter := NewSegmentRewriter(segment_rewrite)
	var dl_interface DL_Interface

	if mode == "mysql" {
		// Fetch list from MySQL.
		log.Infoln("Using mysql as task dispatcher...")
		dl_interface = NewMySQLDl(mysql_host, uint(mysql_port), mysql_db, mysql_table, mysql_username, mysql_password)
	}else if mode == "redis" {
		// Fetch list from Redis.
		log.Infoln("Using redis as task dispatcher...")
		dl_interface = NewRedisDl(redis_host, uint(redis_port), redis_password, redis_db, redis_key)
	}else if flag.NArg() > 0 {
		// Fetch list from Args.
		log.Infoln("Using download list from arguments ...")
		dl_interface = NewDummyDl(flag.Args())
	}else{
		Usage()
		os.Stderr.Write([]byte("\n"))
		return
	}
	hlsgetter := NewHLSGetter(dl_interface, output, path_rewriter, segment_rewriter, retries, timeout, skip, redirect, concurrent, total)
	hlsgetter.SetUA(user_agent)
	hlsgetter.Run()
}