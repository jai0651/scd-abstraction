from .models import Job, Timelog, PaymentLineItem
from .scd.scd_helpers import latest_scd_queryset

def find_active_jobs_by_company(company_id):
    qs = latest_scd_queryset(Job)
    return qs.filter(status='active', company_id=company_id)

def find_active_jobs_by_contractor(contractor_id):
    qs = latest_scd_queryset(Job)
    return qs.filter(status='active', contractor_id=contractor_id)

def find_timelogs_by_contractor_and_period(contractor_id, from_dt, to_dt):
    qs = latest_scd_queryset(Timelog)
    return qs.filter(
        job__contractor_id=contractor_id,
        time_start__gte=from_dt,
        time_end__lte=to_dt
    )

def find_line_items_by_contractor_and_period(contractor_id, from_dt, to_dt):
    qs = latest_scd_queryset(PaymentLineItem)
    return qs.filter(
        job__contractor_id=contractor_id,
        timelog__time_start__gte=from_dt,
        timelog__time_end__lte=to_dt
    )
