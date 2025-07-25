import sys
import os
import django
import datetime
import uuid

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

def uid_for(label, version):
    return f"{label}_v{version}"

def setup_test_data():
    Job.objects.all().delete()
    Timelog.objects.all().delete()
    PaymentLineItem.objects.all().delete()

    # Create Jobs
    job1_v1 = Job.objects.create(
        id='job1', version=1, uid=uid_for('job1', 1),
        status='extended', company_id='comp1', contractor_id='cont1'
    )
    job1_v2 = Job.objects.create(
        id='job1', version=2, uid=uid_for('job1', 2),
        status='active', company_id='comp1', contractor_id='cont1'
    )
    job2 = Job.objects.create(
        id='job2', version=1, uid=uid_for('job2', 1),
        status='active', company_id='comp1', contractor_id='cont2'
    )

    now = datetime.datetime.now()

    # Timelogs
    tl1_v1 = Timelog.objects.create(
        id='tl1', version=1, uid=uid_for('tl1', 1), job=job1_v2,
        time_start=now - datetime.timedelta(hours=2),
        time_end=now - datetime.timedelta(hours=1),
        duration=1.0, type='work'
    )
    tl1_v2 = Timelog.objects.create(
        id='tl1', version=2, uid=uid_for('tl1', 2), job=job1_v2,
        time_start=now - datetime.timedelta(hours=2),
        time_end=now - datetime.timedelta(hours=1),
        duration=1.0, type='work'
    )
    tl2 = Timelog.objects.create(
        id='tl2', version=1, uid=uid_for('tl2', 1), job=job2,
        time_start=now - datetime.timedelta(hours=3),
        time_end=now - datetime.timedelta(hours=2),
        duration=1.0, type='meeting'
    )

    # PaymentLineItems
    PaymentLineItem.objects.create(
        id='pli1', version=1, uid=uid_for('pli1', 1),
        job=job1_v2, timelog=tl1_v2, amount=100.0, status='pending'
    )
    PaymentLineItem.objects.create(
        id='pli1', version=2, uid=uid_for('pli1', 2),
        job=job1_v2, timelog=tl1_v2, amount=120.0, status='paid'
    )

def run_demo():
    setup_test_data()

    company_id = 'comp1'
    contractor_id = 'cont1'
    now = datetime.datetime.now()
    from_dt = now - datetime.timedelta(days=1)
    to_dt = now + datetime.timedelta(days=1)

    print('\n✅ Active jobs for company:')
    for j in find_active_jobs_by_company(company_id):
        print(j)

    print('\n✅ Active jobs for contractor:')
    for j in find_active_jobs_by_contractor(contractor_id):
        print(j)

    print('\n✅ Timelogs for contractor in period:')
    for t in find_timelogs_by_contractor_and_period(contractor_id, from_dt, to_dt):
        print(t)

    print('\n✅ Payment line items for contractor in period:')
    for p in find_line_items_by_contractor_and_period(contractor_id, from_dt, to_dt):
        print(p)

if __name__ == '__main__':
    run_demo()