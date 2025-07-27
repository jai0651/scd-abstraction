#!/usr/bin/env python3
"""
Comparative benchmarking for SCD abstraction in Python
Compares SCD abstraction vs direct SQL vs alternative versioning strategies
"""

import os
import sys
import time
import statistics
import psycopg2
from datetime import datetime, timedelta
from typing import List, Dict, Any
import uuid

# Database connection setup
def get_db_connection():
    return psycopg2.connect(
        host=os.getenv('POSTGRES_HOST', 'localhost'),
        user=os.getenv('POSTGRES_USER', 'postgres'),
        password=os.getenv('POSTGRES_PASSWORD', 'postgres'),
        database=os.getenv('POSTGRES_DB', 'scd_comparative'),
        port=os.getenv('POSTGRES_PORT', '5432')
    )

# SCD Abstraction Implementation (Python equivalent)
class SCDHelper:
    def __init__(self, conn):
        self.conn = conn
    
    def create_new_version(self, table: str, id_value: str, updates: Dict[str, Any]):
        """Create new SCD version using abstraction"""
        cursor = self.conn.cursor()
        
        # Get latest version
        cursor.execute(f"SELECT * FROM {table} WHERE id = %s ORDER BY version DESC LIMIT 1", (id_value,))
        latest = cursor.fetchone()
        
        if not latest:
            raise Exception(f"No record found with id {id_value}")
        
        # Get column names
        cursor.execute(f"SELECT column_name FROM information_schema.columns WHERE table_name = %s ORDER BY ordinal_position", (table,))
        columns = [row[0] for row in cursor.fetchall()]
        
        # Prepare new version data
        new_data = list(latest)
        version_idx = columns.index('version')
        new_data[version_idx] = latest[version_idx] + 1
        
        # Apply updates
        for col, value in updates.items():
            if col in columns:
                col_idx = columns.index(col)
                new_data[col_idx] = value
        
        # Generate new UID
        uid_idx = columns.index('uid')
        new_data[uid_idx] = str(uuid.uuid4())
        
        # Insert new version
        placeholders = ', '.join(['%s'] * len(columns))
        cursor.execute(f"INSERT INTO {table} ({', '.join(columns)}) VALUES ({placeholders})", new_data)
        self.conn.commit()
        cursor.close()
    
    def get_latest_versions(self, table: str) -> List[tuple]:
        """Get latest versions using SCD abstraction"""
        cursor = self.conn.cursor()
        cursor.execute(f"""
            SELECT t1.* FROM {table} t1
            INNER JOIN (
                SELECT id, MAX(version) as max_version 
                FROM {table} 
                GROUP BY id
            ) t2 ON t1.id = t2.id AND t1.version = t2.max_version
        """)
        result = cursor.fetchall()
        cursor.close()
        return result

# Direct SQL Implementation
class DirectSQLHelper:
    def __init__(self, conn):
        self.conn = conn
    
    def create_new_version(self, table: str, id_value: str, updates: Dict[str, Any]):
        """Create new version using direct SQL"""
        cursor = self.conn.cursor()
        
        # Get max version
        cursor.execute(f"SELECT COALESCE(MAX(version), 0) FROM {table} WHERE id = %s", (id_value,))
        max_version = cursor.fetchone()[0]
        new_version = max_version + 1
        
        # Build update clause
        update_clause = ', '.join([f"{k} = %s" for k in updates.keys()])
        update_values = list(updates.values())
        
        # Direct SQL insert with updates
        cursor.execute(f"""
            INSERT INTO {table} (id, version, uid, status, rate, title, company_id, contractor_id)
            SELECT id, %s, %s, 
                   CASE WHEN %s IS NOT NULL THEN %s ELSE status END,
                   CASE WHEN %s IS NOT NULL THEN %s ELSE rate END,
                   title, company_id, contractor_id
            FROM {table} 
            WHERE id = %s AND version = %s
        """, (
            new_version, str(uuid.uuid4()),
            updates.get('status'), updates.get('status'),
            updates.get('rate'), updates.get('rate'),
            id_value, max_version
        ))
        
        self.conn.commit()
        cursor.close()
    
    def get_latest_versions(self, table: str) -> List[tuple]:
        """Get latest versions using direct SQL"""
        cursor = self.conn.cursor()
        cursor.execute(f"""
            SELECT j1.* FROM {table} j1
            INNER JOIN (
                SELECT id, MAX(version) as max_version 
                FROM {table} 
                GROUP BY id
            ) j2 ON j1.id = j2.id AND j1.version = j2.max_version
        """)
        result = cursor.fetchall()
        cursor.close()
        return result

# Alternative Strategy 1: Timestamp-based versioning
class TimestampVersioningHelper:
    def __init__(self, conn):
        self.conn = conn
    
    def create_new_version(self, table: str, id_value: str, updates: Dict[str, Any]):
        """Create new version using timestamp strategy"""
        cursor = self.conn.cursor()
        
        # Get latest version
        cursor.execute(f"SELECT * FROM {table}_ts WHERE id = %s ORDER BY created_at DESC LIMIT 1", (id_value,))
        latest = cursor.fetchone()
        
        if not latest:
            raise Exception(f"No record found with id {id_value}")
        
        # Insert new version with current timestamp
        cursor.execute(f"""
            INSERT INTO {table}_ts (id, created_at, uid, status, rate, title, company_id, contractor_id)
            VALUES (%s, %s, %s, %s, %s, %s, %s, %s)
        """, (
            id_value, datetime.now(), str(uuid.uuid4()),
            updates.get('status', latest[3]),
            updates.get('rate', latest[4]),
            latest[5], latest[6], latest[7]
        ))
        
        self.conn.commit()
        cursor.close()
    
    def get_latest_versions(self, table: str) -> List[tuple]:
        """Get latest versions using timestamp strategy"""
        cursor = self.conn.cursor()
        cursor.execute(f"""
            SELECT t1.* FROM {table}_ts t1
            INNER JOIN (
                SELECT id, MAX(created_at) as max_created 
                FROM {table}_ts 
                GROUP BY id
            ) t2 ON t1.id = t2.id AND t1.created_at = t2.max_created
        """)
        result = cursor.fetchall()
        cursor.close()
        return result

# Alternative Strategy 2: Flag-based versioning
class FlagVersioningHelper:
    def __init__(self, conn):
        self.conn = conn
    
    def create_new_version(self, table: str, id_value: str, updates: Dict[str, Any]):
        """Create new version using flag strategy"""
        cursor = self.conn.cursor()
        
        # Start transaction
        cursor.execute("BEGIN")
        
        # Set current version to false
        cursor.execute(f"UPDATE {table}_flag SET is_current = false WHERE id = %s AND is_current = true", (id_value,))
        
        # Get max version
        cursor.execute(f"SELECT COALESCE(MAX(version), 0) FROM {table}_flag WHERE id = %s", (id_value,))
        max_version = cursor.fetchone()[0]
        
        # Get original data
        cursor.execute(f"SELECT * FROM {table}_flag WHERE id = %s AND version = %s", (id_value, max_version))
        original = cursor.fetchone()
        
        # Insert new version
        cursor.execute(f"""
            INSERT INTO {table}_flag (id, version, uid, is_current, status, rate, title, company_id, contractor_id)
            VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s)
        """, (
            id_value, max_version + 1, str(uuid.uuid4()), True,
            updates.get('status', original[4]),
            updates.get('rate', original[5]),
            original[6], original[7], original[8]
        ))
        
        cursor.execute("COMMIT")
        self.conn.commit()
        cursor.close()
    
    def get_latest_versions(self, table: str) -> List[tuple]:
        """Get latest versions using flag strategy"""
        cursor = self.conn.cursor()
        cursor.execute(f"SELECT * FROM {table}_flag WHERE is_current = true")
        result = cursor.fetchall()
        cursor.close()
        return result

def setup_database():
    """Setup database tables for benchmarking"""
    conn = get_db_connection()
    cursor = conn.cursor()
    
    # Drop existing tables
    tables = ['jobs', 'jobs_ts', 'jobs_flag']
    for table in tables:
        cursor.execute(f"DROP TABLE IF EXISTS {table}")
    
    # Create SCD abstraction table
    cursor.execute("""
        CREATE TABLE jobs (
            id VARCHAR(64),
            version INTEGER,
            uid VARCHAR(64) UNIQUE PRIMARY KEY,
            status VARCHAR(32),
            rate INTEGER,
            title VARCHAR(255),
            company_id VARCHAR(64),
            contractor_id VARCHAR(64),
            PRIMARY KEY (id, version)
        )
    """)
    
    # Create timestamp versioning table
    cursor.execute("""
        CREATE TABLE jobs_ts (
            id VARCHAR(64),
            created_at TIMESTAMP,
            uid VARCHAR(64) UNIQUE PRIMARY KEY,
            status VARCHAR(32),
            rate INTEGER,
            title VARCHAR(255),
            company_id VARCHAR(64),
            contractor_id VARCHAR(64),
            PRIMARY KEY (id, created_at)
        )
    """)
    
    # Create flag versioning table
    cursor.execute("""
        CREATE TABLE jobs_flag (
            id VARCHAR(64),
            version INTEGER,
            uid VARCHAR(64) UNIQUE PRIMARY KEY,
            is_current BOOLEAN,
            status VARCHAR(32),
            rate INTEGER,
            title VARCHAR(255),
            company_id VARCHAR(64),
            contractor_id VARCHAR(64),
            PRIMARY KEY (id, version)
        )
    """)
    
    # Create indexes
    cursor.execute("CREATE INDEX idx_jobs_id_version ON jobs(id, version)")
    cursor.execute("CREATE INDEX idx_jobs_ts_id_created ON jobs_ts(id, created_at)")
    cursor.execute("CREATE INDEX idx_jobs_flag_current ON jobs_flag(is_current)")
    
    conn.commit()
    cursor.close()
    conn.close()

def seed_data(count: int = 1000):
    """Seed test data"""
    conn = get_db_connection()
    cursor = conn.cursor()
    
    # Clear existing data
    cursor.execute("TRUNCATE TABLE jobs, jobs_ts, jobs_flag RESTART IDENTITY CASCADE")
    
    for i in range(count):
        job_id = f"job{i}"
        uid = str(uuid.uuid4())
        
        # SCD abstraction data
        cursor.execute("""
            INSERT INTO jobs (id, version, uid, status, rate, title, company_id, contractor_id)
            VALUES (%s, 1, %s, 'active', 100, 'Engineer', 'comp1', 'cont1')
        """, (job_id, uid))
        
        # Timestamp data
        cursor.execute("""
            INSERT INTO jobs_ts (id, created_at, uid, status, rate, title, company_id, contractor_id)
            VALUES (%s, %s, %s, 'active', 100, 'Engineer', 'comp1', 'cont1')
        """, (job_id, datetime.now(), str(uuid.uuid4())))
        
        # Flag data
        cursor.execute("""
            INSERT INTO jobs_flag (id, version, uid, is_current, status, rate, title, company_id, contractor_id)
            VALUES (%s, 1, %s, true, 'active', 100, 'Engineer', 'comp1', 'cont1')
        """, (job_id, str(uuid.uuid4())))
    
    conn.commit()
    cursor.close()
    conn.close()

def benchmark_function(func, iterations: int = 100) -> Dict[str, float]:
    """Benchmark a function and return timing statistics"""
    times = []
    
    for i in range(iterations):
        start_time = time.time()
        func(i)
        end_time = time.time()
        times.append(end_time - start_time)
    
    return {
        'mean': statistics.mean(times),
        'median': statistics.median(times),
        'stdev': statistics.stdev(times) if len(times) > 1 else 0,
        'min': min(times),
        'max': max(times),
        'total': sum(times)
    }

def run_benchmarks():
    """Run all comparative benchmarks"""
    print("Setting up database...")
    setup_database()
    seed_data(1000)
    
    conn = get_db_connection()
    scd_helper = SCDHelper(conn)
    direct_sql_helper = DirectSQLHelper(conn)
    timestamp_helper = TimestampVersioningHelper(conn)
    flag_helper = FlagVersioningHelper(conn)
    
    results = {}
    iterations = 100
    
    print(f"\nRunning benchmarks with {iterations} iterations each...\n")
    
    # Benchmark version creation
    print("Benchmarking version creation...")
    
    def scd_create_version(i):
        job_id = f"job{i % 1000}"
        scd_helper.create_new_version('jobs', job_id, {'status': 'updated', 'rate': 150})
    
    def direct_sql_create_version(i):
        job_id = f"job{i % 1000}"
        direct_sql_helper.create_new_version('jobs', job_id, {'status': 'updated', 'rate': 150})
    
    def timestamp_create_version(i):
        job_id = f"job{i % 1000}"
        timestamp_helper.create_new_version('jobs', job_id, {'status': 'updated', 'rate': 150})
    
    def flag_create_version(i):
        job_id = f"job{i % 1000}"
        flag_helper.create_new_version('jobs', job_id, {'status': 'updated', 'rate': 150})
    
    results['scd_create'] = benchmark_function(scd_create_version, iterations)
    results['direct_sql_create'] = benchmark_function(direct_sql_create_version, iterations)
    results['timestamp_create'] = benchmark_function(timestamp_create_version, iterations)
    results['flag_create'] = benchmark_function(flag_create_version, iterations)
    
    # Benchmark latest version queries
    print("Benchmarking latest version queries...")
    
    def scd_query_latest(i):
        scd_helper.get_latest_versions('jobs')
    
    def direct_sql_query_latest(i):
        direct_sql_helper.get_latest_versions('jobs')
    
    def timestamp_query_latest(i):
        timestamp_helper.get_latest_versions('jobs')
    
    def flag_query_latest(i):
        flag_helper.get_latest_versions('jobs')
    
    results['scd_query'] = benchmark_function(scd_query_latest, iterations)
    results['direct_sql_query'] = benchmark_function(direct_sql_query_latest, iterations)
    results['timestamp_query'] = benchmark_function(timestamp_query_latest, iterations)
    results['flag_query'] = benchmark_function(flag_query_latest, iterations)
    
    conn.close()
    
    # Print results
    print("\n" + "="*80)
    print("PYTHON BENCHMARK RESULTS")
    print("="*80)
    
    print("\nVERSION CREATION PERFORMANCE (seconds):")
    print(f"{'Strategy':<20} {'Mean':<10} {'Median':<10} {'StdDev':<10} {'Min':<10} {'Max':<10}")
    print("-" * 70)
    
    for strategy in ['scd_create', 'direct_sql_create', 'timestamp_create', 'flag_create']:
        stats = results[strategy]
        name = strategy.replace('_create', '').replace('_', ' ').title()
        print(f"{name:<20} {stats['mean']:<10.6f} {stats['median']:<10.6f} {stats['stdev']:<10.6f} {stats['min']:<10.6f} {stats['max']:<10.6f}")
    
    print("\nLATEST VERSION QUERY PERFORMANCE (seconds):")
    print(f"{'Strategy':<20} {'Mean':<10} {'Median':<10} {'StdDev':<10} {'Min':<10} {'Max':<10}")
    print("-" * 70)
    
    for strategy in ['scd_query', 'direct_sql_query', 'timestamp_query', 'flag_query']:
        stats = results[strategy]
        name = strategy.replace('_query', '').replace('_', ' ').title()
        print(f"{name:<20} {stats['mean']:<10.6f} {stats['median']:<10.6f} {stats['stdev']:<10.6f} {stats['min']:<10.6f} {stats['max']:<10.6f}")
    
    # Performance comparison
    print("\nPERFORMANCE COMPARISON:")
    print("-" * 40)
    
    scd_create_time = results['scd_create']['mean']
    direct_sql_create_time = results['direct_sql_create']['mean']
    
    if scd_create_time < direct_sql_create_time:
        improvement = ((direct_sql_create_time - scd_create_time) / direct_sql_create_time) * 100
        print(f"SCD Abstraction is {improvement:.1f}% faster than Direct SQL for version creation")
    else:
        overhead = ((scd_create_time - direct_sql_create_time) / direct_sql_create_time) * 100
        print(f"SCD Abstraction has {overhead:.1f}% overhead compared to Direct SQL for version creation")
    
    scd_query_time = results['scd_query']['mean']
    direct_sql_query_time = results['direct_sql_query']['mean']
    
    if scd_query_time < direct_sql_query_time:
        improvement = ((direct_sql_query_time - scd_query_time) / direct_sql_query_time) * 100
        print(f"SCD Abstraction is {improvement:.1f}% faster than Direct SQL for queries")
    else:
        overhead = ((scd_query_time - direct_sql_query_time) / direct_sql_query_time) * 100
        print(f"SCD Abstraction has {overhead:.1f}% overhead compared to Direct SQL for queries")
    
    return results

if __name__ == "__main__":
    run_benchmarks()
