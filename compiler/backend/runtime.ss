(define %methods #f)
(define %resume-type-offset #f)

(define (method-invoke object name . args)
  (apply (method-lookup object name) object args))

(define (method-lookup object name)
  (let* ([class-id (vector-ref object 0)]
         [offset (+ class-id name)])
    (and (method-valid? name offset)
         (method-impl offset))))

(define (method-valid? name offset)
  (and (< (* offset 2) (vector-length %methods))
       (= name (method-name offset))))

(define (method-name idx) (vector-ref %methods (* idx 2)))
(define (method-impl idx) (vector-ref %methods (+ (* idx 2) 1)))

(define-record-type context 
  (fields handlers continuation))

(define %waiting           '())
(define %handlers          #f)
(define %base-continuation #f)

(define (run prog)
  (let ([x ((call/cc (lambda (k) (set! %base-continuation k) prog)))])
    (if (null? %waiting)
      x
      (let ([next (car %waiting)])
        (set! %waiting (cdr %waiting))
        (set! %handlers (context-handlers next))
        ((context-continuation next) x)))))

(define-syntax define/cc 
  (syntax-rules ()
    [(_ (name k . args) body ...)
     (define (name . args)
       (call/cc (lambda (k) body ...)))]))

(define/cc (install-handlers k hs prog)
  (let ([next (make-context %handlers k)])
    (set! %waiting (cons next %waiting))
    (set! %handlers hs)
    (%base-continuation prog)))

(define/cc (trigger-effect k name . args)
  (let ([inner (cons (make-context %handlers k) %waiting)])
    (let next ([curr inner])
      (let* ([handlers (context-handlers (car curr))]
             [handler (method-lookup handlers name)])
        (if handler
            (let ([resume (make-resumer handlers inner (cdr curr))])
              (set! %waiting (cdr curr))
              (%base-continuation (lambda () (apply handler handlers resume args))))
            (next (cdr curr)))))))

(define (make-resumer handlers start stop)
  (vector %resume-type-offset handlers start stop))

(define/cc (resumer:resume k this x)
  (let ([handlers (vector-ref this 1)]
        [start (vector-ref this 2)]
        [stop (vector-ref this 3)])
    (set! %waiting (let next ([curr start])
                    (if (eqv? curr stop)
                        (cons (make-context handlers k) %waiting)
                        (cons (car curr) (next (cdr curr))))))
    (%base-continuation (lambda () x))))

(define (arith:add this x y) (+ x y))
(define (sys:print this x) (display x) (newline))
