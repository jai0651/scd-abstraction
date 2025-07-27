import sys
import os
import django
import datetime
import time

# Set up Django
os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'myproject.settings')
sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), '..')))
django.setup()

from django_example.models import Job, Timelog, PaymentLineItem
from django_example.repos import (
    find_active_jobs_by_company,
    find_active_jobs_by_contractor,
    find_timelogs_by_contractor_and_period,
    find_line_items_by_contractor_and_period,
)

def seed_data(n=100000):
    print(f"Seeding {n} jobs, timelogs, and payment line items...")
    Job.objects.all().delete()
    Timelog.objects.all().delete()
    PaymentLineItem.objects.all().delete()
    jobs = []
    timelogs = []
    plis = []
    now = datetime.datetime.now()
    for i in range(n):
        job_uid = f"job-uid-{i}"
        job = Job(
            uid=job_uid,
            id=f"job{i}",
            version=1,
            status="active",
            rate=100,
            title="Engineer",
            company_id="comp1",
            contractor_id="cont1"
        )
        jobs.append(job)
    Job.objects.bulk_create(jobs, batch_size=1000)
    jobs = {job.uid: job for job in Job.objects.filter(company_id="comp1")}
    for i in range(n):
        job_uid = f"job-uid-{i}"
        timelog_uid = f"tl-uid-{i}"
        timelog = Timelog(
            uid=timelog_uid,
            id=f"tl{i}",
            version=1,
            duration=8,
            time_start=now - datetime.timedelta(hours=2),
            time_end=now - datetime.timedelta(hours=1),
            type="work",
            job=jobs[job_uid]
        )
        timelogs.append(timelog)
    Timelog.objects.bulk_create(timelogs, batch_size=1000)
    timelogs = {tl.uid: tl for tl in Timelog.objects.filter(job__company_id="comp1")}
    for i in range(n):
        job_uid = f"job-uid-{i}"
        timelog_uid = f"tl-uid-{i}"
        pli = PaymentLineItem(
            uid=f"pli-uid-{i}",
            id=f"pli{i}",
            version=1,
            job=jobs[job_uid],
            timelog=timelogs[timelog_uid],
            amount=800,
            status="pending"
        )
        plis.append(pli)
    PaymentLineItem.objects.bulk_create(plis, batch_size=1000)
    print("Seeding complete.")

def benchmark_query(label, func, *args):
    start = time.time()
    result = list(func(*args))
    elapsed = time.time() - start
    print(f"{label}: {elapsed:.4f}s, {len(result)} results")
    return elapsed

if __name__ == "__main__":
    seed_data(100000)
    company_id = "comp1"
    contractor_id = "cont1"
    now = datetime.datetime.now()
    from_dt = now - datetime.timedelta(days=1)
    to_dt = now + datetime.timedelta(days=1)

    print("\nBenchmarking queries...")
    benchmark_query("FindActiveJobsByCompany", find_active_jobs_by_company, company_id)
    benchmark_query("FindActiveJobsByContractor", find_active_jobs_by_contractor, contractor_id)
    benchmark_query("FindTimelogsByContractorAndPeriod", find_timelogs_by_contractor_and_period, contractor_id, from_dt, to_dt)
    benchmark_query("FindLineItemsByContractorAndPeriod", find_line_items_by_contractor_and_period, contractor_id, from_dt, to_dt) 