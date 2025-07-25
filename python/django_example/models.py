from django.db import models


class Job(models.Model):
    uid = models.CharField(max_length=64, unique=True, primary_key=True)
    id = models.CharField(max_length=64, db_index=True)
    version = models.IntegerField()
    status = models.CharField(max_length=32)
    rate = models.FloatField()
    title = models.CharField(max_length=255)
    company_id = models.CharField(max_length=64)
    contractor_id = models.CharField(max_length=64)

    class Meta:
        db_table = 'jobs'

    def __str__(self):
        return f"{self.title} ({self.uid})"


class Timelog(models.Model):
    uid = models.CharField(max_length=64, unique=True, primary_key=True)
    id = models.CharField(max_length=64, db_index=True)
    version = models.IntegerField()
    duration = models.FloatField()
    time_start = models.DateTimeField()
    time_end = models.DateTimeField()
    type = models.CharField(max_length=32)
    job = models.ForeignKey(
        Job,
        on_delete=models.CASCADE,
        to_field='uid',
        db_column='job_uid',
        related_name='timelogs'
    )
  

    class Meta:
        db_table = 'timelogs'

    def __str__(self):
        return f"Timelog {self.uid}"


class PaymentLineItem(models.Model):
    uid = models.CharField(max_length=64, unique=True, primary_key=True)
    id = models.CharField(max_length=64, db_index=True)
    version = models.IntegerField()
    job = models.ForeignKey(
        Job,
        on_delete=models.CASCADE,
        to_field='uid',
        db_column='job_uid',
        related_name='payment_line_items'
    )
    timelog = models.ForeignKey(
        Timelog,
        on_delete=models.CASCADE,
        to_field='uid',
        db_column='timelog_uid',
        related_name='payment_line_items'
    )
    amount = models.FloatField()
    status = models.CharField(max_length=32)
    
    class Meta:
        db_table = 'payment_line_items'

    def __str__(self):
        return f"PaymentLineItem {self.uid}"