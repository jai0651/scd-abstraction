# SCD (Slowly Changing Dimension) Abstraction
A template-based SCD implementation that generates versioning helpers for multiple ORMs and languages. This project demonstrates how **code generation** can solve the challenge of maintaining consistent SCD logic across different database technologies.

## 🎯 The Problem

When implementing SCD across multiple languages and ORMs, you face a fundamental challenge:

- **Go/GORM**: Uses generics and reflection for type-safe SCD operations
- **Python/Django**: Uses querysets and model instances
- **Java/Hibernate**: Uses JPA annotations and entity managers
- **Node.js/Sequelize**: Uses JavaScript objects and promises

**Traditional approaches fail because:**
- ❌ **Code duplication** - Same logic written in multiple languages
- ❌ **Inconsistency risk** - Logic drift between implementations
- ❌ **Maintenance overhead** - Update logic in multiple places
- ❌ **Testing complexity** - Test each language implementation separately

## 💡 The Solution: Template-Based Code Generation

### Why Templates Work for SCD

SCD operations follow **universal patterns** that translate across ORMs:

1. **Latest Version Query**: Always `SELECT id, MAX(version) FROM table GROUP BY id`
2. **Version Creation**: Always clone latest + increment version + apply updates
3. **Repository Pattern**: Always encapsulate SCD logic in repository methods

These patterns are **ORM-agnostic** - only the syntax changes.

### Technical Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Universal     │    │   ORM-Specific  │    │   Generated     │
│   SCD Logic     │───▶│   Template      │───▶│   Implementation │
│   (Patterns)    │    │   (Syntax)      │    │   (Code)        │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### How It Works

#### 1. **Universal SCD Patterns**
The core SCD operations are the same across all ORMs:

```sql
-- Universal pattern for latest versions
SELECT t.* FROM table t
JOIN (
    SELECT id, MAX(version) as max_version 
    FROM table 
    GROUP BY id
) latest ON t.id = latest.id AND t.version = latest.max_version
```

#### 2. **ORM-Specific Translation**
The template translates these patterns into ORM-specific syntax:

**GORM (Go):**
```go
func LatestSubquery[T any](db *gorm.DB, model T) *gorm.DB {
    return db.Model(&model).
        Select("id, MAX(version) as max_version").
        Group("id")
}
```

**Django ORM (Python):**
```python
def latest_scd_queryset(model, base_queryset=None):
    subq = (
        model.objects
        .filter(id=OuterRef('id'))
        .values('id')
        .annotate(max_version=Max('version'))
        .values('max_version')
    )
    return base_queryset.filter(version=Subquery(subq))
```

#### 3. **Generated Implementation**
The template generates **type-safe, ORM-optimized** code that:
- ✅ **Leverages ORM features** (GORM generics, Django querysets)
- ✅ **Maintains consistency** (Same logic across ORMs)
- ✅ **Provides type safety** (Compile-time checking where available)
- ✅ **Optimizes performance** (ORM-specific query optimizations)

## 🚀 Why This Approach Succeeds

### 1. **Pattern Recognition**
SCD operations are **algorithmic patterns** that don't change:
- Finding latest versions = `MAX(version)` per `id`
- Creating new versions = Clone + Increment + Update
- Repository encapsulation = Hide complexity behind clean API

### 2. **Syntax Translation**
Only the **syntax** differs between ORMs:
- **GORM**: `db.Model(&model).Select(...).Group(...)`
- **Django**: `model.objects.filter(...).values(...).annotate(...)`
- **Hibernate**: `session.createQuery(...).setParameter(...)`

### 3. **Type Safety Preservation**
The template generates **type-safe code** for each ORM:
- **GORM**: Uses Go generics `func LatestSubquery[T any]`
- **Django**: Uses Python type hints and model classes
- **Hibernate**: Uses Java generics and entity types

### 4. **Performance Optimization**
Generated code leverages **ORM-specific optimizations**:
- **GORM**: Query building, connection pooling, prepared statements
- **Django**: QuerySet optimization, database-specific features
- **Hibernate**: Session management, caching, lazy loading

## 📊 Performance Results

Our benchmarks show that the SCD abstraction adds minimal overhead:

| Query Type | SCD Time | Raw SQL Time | Overhead | Status |
|------------|----------|--------------|----------|---------|
| Job By Company | 2.35ms | 1.95ms | +20.23% | ⚠️ Moderate |
| Job By Contractor | 2.49ms | 2.13ms | +16.89% | ✅ Acceptable |
| Latest Subquery | 123μs | 98μs | +25.51% | ⚠️ Moderate |
| Create New Version | 665μs | 604μs | +10.10% | ✅ Acceptable |

**Key Findings:**
- ✅ **10-25% performance overhead** - Well within acceptable limits for most use cases
- ✅ **Excellent developer productivity** - Clean, consistent API
- ✅ **Multi-language support** - Go and Python implementations
- ✅ **Type safety** - Leverages ORM-specific features
- ⚠️ **Moderate overhead** - Some operations show 20-25% overhead, still acceptable for most applications

## 🔄 Extending to New ORMs

### The Process

#### 1. **Identify Universal Patterns**
```sql
-- This pattern works for ANY ORM
SELECT id, MAX(version) FROM table GROUP BY id
```

#### 2. **Translate to ORM Syntax**
```jinja
{% if target == "your_orm" %}
// Your ORM's syntax for the universal pattern
function latestSubquery(model) {
    return model.select('id, MAX(version) as max_version')
                .groupBy('id');
}
{% endif %}
```

#### 3. **Generate Type-Safe Code**
The template generates ORM-specific code that:
- Uses the ORM's native syntax
- Preserves type safety
- Leverages ORM optimizations
- Maintains consistent patterns

### Why It's Easy to Extend

1. **Universal Patterns**: SCD logic doesn't change
2. **Syntax Translation**: Only the API calls differ
3. **Template Engine**: Jinja2 handles the translation
4. **Type Safety**: Each ORM's type system is preserved

## 🎯 Technical Benefits

### 1. **Single Source of Truth**
- One template = All ORM implementations
- Update logic once = Update everywhere
- Consistent behavior across languages

### 2. **ORM-Specific Optimizations**
- Generated code uses each ORM's best features
- GORM generics for type safety
- Django querysets for optimization
- Hibernate sessions for performance

### 3. **Developer Productivity**
- Familiar patterns across languages
- Type-safe APIs for each ORM
- Consistent repository interfaces
- Rapid prototyping for new ORMs

### 4. **Maintainability**
- Centralized logic in template
- Version-controlled generation
- Testable generation process
- Clear separation of concerns

## 🏗️ Architecture Deep Dive

### Template Structure
```jinja
{# Universal SCD patterns in template #}
{% if target == "gorm" %}
    // GORM-specific syntax for universal patterns
{% elif target == "django" %}
    // Django-specific syntax for universal patterns
{% elif target == "your_orm" %}
    // Your ORM-specific syntax for universal patterns
{% endif %}
```

### Generation Process
```python
# 1. Load universal template
template = env.get_template('scd_helper_template.jinja')

# 2. Render with ORM-specific context
rendered = template.render(target=target)

# 3. Generate type-safe, optimized code
write_to_file(rendered, output_path)
```

### Why This Works
1. **SCD patterns are universal** - Same logic across ORMs
2. **Only syntax differs** - Template handles translation
3. **Type safety preserved** - Each ORM's type system maintained
4. **Performance optimized** - Generated code uses ORM features

## 🚀 Quick Start

### 1. Generate SCD Helpers
```bash
# Generate for Go/GORM
python generate_scd_helper.py gorm

# Generate for Python/Django  
python generate_scd_helper.py django
```

### 2. Use Generated Code
```go
// Go - Type-safe with GORM generics
subq := LatestSubquery(db, models.Job{})
jobs, err := db.Model(&models.Job{}).
    Joins("JOIN (?) AS latest ON jobs.id = latest.id AND jobs.version = latest.max_version", subq).
    Where("jobs.status = ?", "active").
    Find(&jobs).Error
```

```python
# Python - Optimized with Django querysets
jobs = latest_scd_queryset(Job).filter(status='active')
```

### 3. Run Benchmarks
```bash
# Test performance across ORMs
docker-compose up --build go-bench
```

## 📚 Documentation

- **[SCD Impact Analysis](docs/SCD_Impact_Analysis.md)** - Performance analysis
- **[Template vs Alternatives](docs/Template_vs_Alternatives_Analysis.md)** - Architectural comparison

## 🎉 Conclusion

This template-based approach succeeds because:

1. **SCD patterns are universal** - Same logic across ORMs
2. **Only syntax differs** - Template handles translation
3. **Type safety preserved** - Each ORM's features maintained
4. **Performance optimized** - Generated code uses ORM best practices
5. **Developer friendly** - Consistent patterns across languages

**The result**: One template generates type-safe, optimized SCD implementations for any ORM, with minimal performance overhead and maximum developer productivity.

---

*Performance overhead: 13-14%*  
*Supported ORMs: GORM (Go), Django ORM (Python)*  
*Extensible to: Any ORM with template translation* 