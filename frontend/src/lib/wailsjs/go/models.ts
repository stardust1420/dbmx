export namespace model {
	
	export class Cell {
	    column: string;
	    value: string;
	
	    static createFrom(source: any = {}) {
	        return new Cell(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.column = source["column"];
	        this.value = source["value"];
	    }
	}
	export class Connection {
	    ID: number;
	    Engine: string;
	    Host: string;
	    Port: string;
	    Username: string;
	    Password: string;
	    Database: string;
	    Name: string;
	    Env: string;
	    Color: string;
	    IsAdvanced: boolean;
	    SSLMode: string;
	    ClientKey: number[];
	    ClientCert: number[];
	    RootCACert: number[];
	    OverSSH: boolean;
	    SSHHost: string;
	    SSHPort: string;
	    SSHUsername: string;
	    SSHPassword: string;
	    UseSSHKey: boolean;
	    SSHKey: number[];
	    IsActive: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Connection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Engine = source["Engine"];
	        this.Host = source["Host"];
	        this.Port = source["Port"];
	        this.Username = source["Username"];
	        this.Password = source["Password"];
	        this.Database = source["Database"];
	        this.Name = source["Name"];
	        this.Env = source["Env"];
	        this.Color = source["Color"];
	        this.IsAdvanced = source["IsAdvanced"];
	        this.SSLMode = source["SSLMode"];
	        this.ClientKey = source["ClientKey"];
	        this.ClientCert = source["ClientCert"];
	        this.RootCACert = source["RootCACert"];
	        this.OverSSH = source["OverSSH"];
	        this.SSHHost = source["SSHHost"];
	        this.SSHPort = source["SSHPort"];
	        this.SSHUsername = source["SSHUsername"];
	        this.SSHPassword = source["SSHPassword"];
	        this.UseSSHKey = source["UseSSHKey"];
	        this.SSHKey = source["SSHKey"];
	        this.IsActive = source["IsActive"];
	    }
	}
	export class Database {
	    ID: string;
	    ConnectionID: number;
	    ConnectionName: string;
	    Name: string;
	    Color: string;
	    PoolID: string;
	    IsActive: boolean;
	    Tables: string[];
	    Columns: string[];
	
	    static createFrom(source: any = {}) {
	        return new Database(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.ConnectionID = source["ConnectionID"];
	        this.ConnectionName = source["ConnectionName"];
	        this.Name = source["Name"];
	        this.Color = source["Color"];
	        this.PoolID = source["PoolID"];
	        this.IsActive = source["IsActive"];
	        this.Tables = source["Tables"];
	        this.Columns = source["Columns"];
	    }
	}
	export class Indexes {
	    columns: string[];
	    rows: Cell[][];
	
	    static createFrom(source: any = {}) {
	        return new Indexes(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.columns = source["columns"];
	        this.rows = this.convertValues(source["rows"], Cell);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Output {
	    columns: string[];
	    rows: Cell[][];
	
	    static createFrom(source: any = {}) {
	        return new Output(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.columns = source["columns"];
	        this.rows = this.convertValues(source["rows"], Cell);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class QueryResult {
	    ok: boolean;
	    columns: string[];
	    rows: Cell[][];
	    rowsAffected: number;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new QueryResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.columns = source["columns"];
	        this.rows = this.convertValues(source["rows"], Cell);
	        this.rowsAffected = source["rowsAffected"];
	        this.message = source["message"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Rules {
	    columns: string[];
	    rows: Cell[][];
	
	    static createFrom(source: any = {}) {
	        return new Rules(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.columns = source["columns"];
	        this.rows = this.convertValues(source["rows"], Cell);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Structure {
	    columns: string[];
	    rows: Cell[][];
	
	    static createFrom(source: any = {}) {
	        return new Structure(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.columns = source["columns"];
	        this.rows = this.convertValues(source["rows"], Cell);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Tab {
	    ID: number;
	    Name: string;
	    Editor: string;
	    Output: string;
	    IsActive: boolean;
	    ActiveDBID?: string;
	    ActiveDB?: string;
	    ActiveDBColor?: string;
	    Type: string;
	    columns: string[];
	    rows: Cell[][];
	    PostgresConnID?: number;
	    PostgresConnName: string;
	    DBName?: string;
	    Select: string;
	    Limit: string;
	    Offset: string;
	    Where: string;
	    OrderBy: string;
	    GroupBy: string;
	    TableColumns: string;
	    TableColumnsList: string[];
	
	    static createFrom(source: any = {}) {
	        return new Tab(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	        this.Editor = source["Editor"];
	        this.Output = source["Output"];
	        this.IsActive = source["IsActive"];
	        this.ActiveDBID = source["ActiveDBID"];
	        this.ActiveDB = source["ActiveDB"];
	        this.ActiveDBColor = source["ActiveDBColor"];
	        this.Type = source["Type"];
	        this.columns = source["columns"];
	        this.rows = this.convertValues(source["rows"], Cell);
	        this.PostgresConnID = source["PostgresConnID"];
	        this.PostgresConnName = source["PostgresConnName"];
	        this.DBName = source["DBName"];
	        this.Select = source["Select"];
	        this.Limit = source["Limit"];
	        this.Offset = source["Offset"];
	        this.Where = source["Where"];
	        this.OrderBy = source["OrderBy"];
	        this.GroupBy = source["GroupBy"];
	        this.TableColumns = source["TableColumns"];
	        this.TableColumnsList = source["TableColumnsList"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class TableInfo {
	    structure: Structure;
	    indexes: Indexes;
	    rules: Rules;
	
	    static createFrom(source: any = {}) {
	        return new TableInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.structure = this.convertValues(source["structure"], Structure);
	        this.indexes = this.convertValues(source["indexes"], Indexes);
	        this.rules = this.convertValues(source["rules"], Rules);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace pgx {
	
	export class ConnConfig {
	    Host: string;
	    Port: number;
	    Database: string;
	    User: string;
	    Password: string;
	    // Go type: tls
	    TLSConfig?: any;
	    ConnectTimeout: number;
	    RuntimeParams: Record<string, string>;
	    KerberosSrvName: string;
	    KerberosSpn: string;
	    Fallbacks: pgconn.FallbackConfig[];
	    Tracer: any;
	    StatementCacheCapacity: number;
	    DescriptionCacheCapacity: number;
	    DefaultQueryExecMode: number;
	
	    static createFrom(source: any = {}) {
	        return new ConnConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Host = source["Host"];
	        this.Port = source["Port"];
	        this.Database = source["Database"];
	        this.User = source["User"];
	        this.Password = source["Password"];
	        this.TLSConfig = this.convertValues(source["TLSConfig"], null);
	        this.ConnectTimeout = source["ConnectTimeout"];
	        this.RuntimeParams = source["RuntimeParams"];
	        this.KerberosSrvName = source["KerberosSrvName"];
	        this.KerberosSpn = source["KerberosSpn"];
	        this.Fallbacks = this.convertValues(source["Fallbacks"], pgconn.FallbackConfig);
	        this.Tracer = source["Tracer"];
	        this.StatementCacheCapacity = source["StatementCacheCapacity"];
	        this.DescriptionCacheCapacity = source["DescriptionCacheCapacity"];
	        this.DefaultQueryExecMode = source["DefaultQueryExecMode"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

